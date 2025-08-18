#!/usr/bin/env tsx

import { readFileSync, existsSync } from 'fs';
import { join, dirname, basename, extname } from 'path';
import { fileURLToPath } from 'url';
import { glob } from 'glob';
import { gzipSync } from 'zlib';
import Arweave from 'arweave';
import { JWKInterface } from 'arweave/node/lib/wallet';
import chalk from 'chalk';
import ora from 'ora';
import { ANT, AOProcess, ARIO, ARIO_MAINNET_PROCESS_ID, ArweaveSigner } from '@ar.io/sdk';
import { TurboAuthenticatedClient, TurboFactory, TurboUploadDataItemResponse } from '@ardrive/turbo-sdk';
import { connect } from '@permaweb/aoconnect';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Types
interface Config {
  arweave: {
    host: string;
    port: number;
    protocol: string;
  };
  network: {
    gateway: string;
    cuUrl: string;
  },
  arns: {
    undername: string;
    name: string;
    registry: string;
  };
  paths: {
    dist: string;
    wallet: string;
  };
  dryRun: boolean;
}

interface PlatformArch {
  platform: string;
  arch: string;
}

interface ManifestAsset {
  platform: string;
  arch: string;
  url: string;
  arweave_id: string;
}

interface ReleasesData {
  releases: Array<{
    tag_name: string;
    version: string;
    created_at: string;
    assets: ManifestAsset[];
  }>;
}

interface ArweaveManifest {
  manifest: string;
  version: string;
  index: {
    path: string;
  };
  paths: Record<string, { id: string }>;
}

interface ManifestResult {
  manifestId: string;
  releasesId: string;
  manifest: ArweaveManifest;
}

// Parse command line arguments
const args = process.argv.slice(2);
const isDryRun = args.includes('--dryrun') || args.includes('--dry-run') || process.env.DRYRUN === 'true';

const dataItemOptions = {
    tags: [
        {name: 'Type', value: 'release'},
        {name: 'App-Name', value: 'Harlequin-CLI'},
    ]
}

// Configuration
const CONFIG: Config = {
  arweave: {
    host: 'arweave.net',
    port: 443,
    protocol: 'https'
  },
  network: {
    gateway: process.env.ARWEAVE_GATEWAY || 'https://arweave.net',
    cuUrl: process.env.ARWEAVE_CU_URL || 'https://cu.ardrive.io'
  },
  arns: {
    undername: process.env.ARNS_UNDERNAME || 'install_cli',
    name: process.env.ARNS_NAME || 'harlequin',
    registry: process.env.ARNS_REGISTRY || ARIO_MAINNET_PROCESS_ID
  },
  paths: {
    dist: join(__dirname, '../../dist'),
    wallet: process.env.ARWEAVE_WALLET_PATH || join(process.env.HOME || '', '.arweave-wallet.json')
  },
  dryRun: isDryRun
};

// Initialize Arweave
const arweave = Arweave.init(CONFIG.arweave);

// Initialize turbo client
const turbo = TurboFactory.unauthenticated();

/**
 * Load wallet from file or environment
 */
async function loadWallet(): Promise<JWKInterface> {
  const prefix = CONFIG.dryRun ? '[DRYRUN]' : '';
  const spinner = ora(`${prefix} Loading Arweave wallet...`.trim()).start();
  
  try {
    if (CONFIG.dryRun) {
      spinner.succeed(`${prefix} Mock wallet loaded (dry run mode)`);
      const w = await arweave.wallets.generate();
      return w; // Return in memory wallet for dry run
    }
    
    let walletData: JWKInterface;
    
    if (process.env.ARWEAVE_WALLET_JWK) {
      // Load from environment variable (for CI/CD)
      walletData = JSON.parse(process.env.ARWEAVE_WALLET_JWK);
    } else if (existsSync(CONFIG.paths.wallet)) {
      // Load from file (for local development)
      walletData = JSON.parse(readFileSync(CONFIG.paths.wallet, 'utf8'));
    } else {
      throw new Error('No wallet found. Set ARWEAVE_WALLET_JWK env var or wallet file path.');
    }
    
    const address = await arweave.wallets.jwkToAddress(walletData);
    const arBalance = await arweave.wallets.getBalance(address).catch(() => '0');
    const turboBalance = await turbo.getBalance(address).catch(() => ({winc: '0'}));
    
    spinner.succeed(`Wallet loaded: ${address} (${arweave.ar.winstonToAr(arBalance)} AR, ${arweave.ar.winstonToAr(turboBalance.winc)} TURBO Credits)`);
    return walletData;
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    spinner.fail(`Failed to load wallet: ${errorMessage}`);
    throw error;
  }
}

/**
 * Get CLI version from package.json or environment
 */
function getVersion(): string {
  const version = process.env.CLI_VERSION || 
                 process.env.GITHUB_REF_NAME?.replace('cli-v', '') ||
                 '0.0.0-dev';
  
  console.log(chalk.blue(`📦 Deploying CLI version: ${version}`));
  return version;
}

/**
 * Find all binary files in dist directory
 */
function findBinaries(): string[] {
  const spinner = ora('Finding binary files...').start();
  
  try {
    if (!existsSync(CONFIG.paths.dist)) {
      throw new Error(`Dist directory not found: ${CONFIG.paths.dist}`);
    }
    
    // Find all files (archives and extracted binaries)
    const files = glob.sync('**/*', {
      cwd: CONFIG.paths.dist,
      nodir: true,
      absolute: true
    });
    
    const binaries = files.filter(file => {
      const fileName = basename(file);
      const ext = extname(file);
      
      // Include archives and executable binaries
      return ext === '.tar.gz' || 
             ext === '.zip' || 
             fileName === 'harlequin' || 
             fileName === 'harlequin.exe' ||
             fileName.includes('harlequin');
    });
    
    spinner.succeed(`Found ${binaries.length} binary files`);
    return binaries;
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    spinner.fail(`Failed to find binaries: ${errorMessage}`);
    throw error;
  }
}

/**
 * Parse platform and architecture from filename
 */
function parsePlatformArch(filename: string): PlatformArch {
  const patterns = [
    // Archive format: harlequin_1.2.3_linux_amd64.tar.gz
    /harlequin_[\d.]+_([^_]+)_([^.]+)/,
    // Binary format: harlequin-linux-amd64
    /harlequin-([^-]+)-([^.-]+)/
  ];
  
  for (const pattern of patterns) {
    const match = filename.match(pattern);
    if (match) {
      return {
        platform: match[1]!,
        arch: match[2]!
      };
    }
  }
  
  // Default fallback
  if (filename.includes('linux')) return { platform: 'linux', arch: 'amd64' };
  if (filename.includes('darwin')) return { platform: 'darwin', arch: 'amd64' };
  if (filename.includes('windows')) return { platform: 'windows', arch: 'amd64' };
  
  return { platform: 'unknown', arch: 'unknown' };
}

function pathsFromReleases(releasesData: ReleasesData): Record<string, { id: string }> {
  const paths: Record<string, { id: string }> = {};
  for (const release of releasesData.releases) {
    for (const asset of release.assets) {
      paths[`/releases/${release.version}/${asset.platform}/${asset.arch}`] = { id: asset.arweave_id };
    }
  }
  return paths;
}

/**
 * Create Arweave manifest for routing
 */
async function createManifest(turboClient: TurboAuthenticatedClient, uploads: {uploadResponse: TurboUploadDataItemResponse, name: string}[], installScript: TurboUploadDataItemResponse | null): Promise<ManifestResult> {
  const prefix = CONFIG.dryRun ? '[DRYRUN]' : '';
  const spinner = ora(`${prefix} Creating Arweave manifest...`.trim()).start();
  
  try {
    // load existing releases, return empty array if error
    const existingReleases = await fetch(`https://${CONFIG.arns.name}.arweave.net/releases`).then(res => res.json() as Promise<ReleasesData>).catch(err => {
        console.error(chalk.red('Failed to fetch existing releases:'));
        console.error(chalk.red(err.message));
        return {releases: []};
    });

    const version = getVersion();
    const manifest: ArweaveManifest = {
      manifest: 'arweave/paths',
      version: '0.1.0',
      index: {
        path: 'install_cli.sh'
      },
      paths: pathsFromReleases(existingReleases)
    };
    
    // Add install script
    if (installScript) {
      manifest.paths['install_cli.sh'] = {
        id: installScript.id
      };
    }
    
    // Add releases API endpoint
    const releasesData: ReleasesData = {
      releases: [
        {
          tag_name: `harlequin-cli-v${version}`,
          version: version,
          created_at: new Date().toISOString(),
          assets: []
        }
      ]
    };
    
    // Process uploads and add to manifest
    uploads.forEach(upload => {
      const { platform, arch } = parsePlatformArch(upload.name);
      const uploadResponse = upload.uploadResponse;
      
      if (platform !== 'unknown') {
        // Add to releases API data
        releasesData.releases[0]!.assets.push({
          platform,
          arch,
          url: `https://${CONFIG.arns.name}.arweave.net/releases/${version}/${platform}/${arch}`,
          arweave_id: uploadResponse.id
        });
        
        // Add to manifest paths
        manifest.paths[`releases/${version}/${platform}/${arch}`] = {
          id: uploadResponse.id
        };
        
        // Add latest symlinks
        manifest.paths[`releases/latest/${platform}/${arch}`] = {
          id: uploadResponse.id
        };
      }
    });
    
    if (CONFIG.dryRun) {
      // Mock IDs for dry run
      const mockReleasesId = `mock_releases_${Math.random().toString(36).substr(2, 43)}`;
      const mockManifestId = `mock_manifest_${Math.random().toString(36).substr(2, 43)}`;
      
      spinner.succeed(`${prefix} Would create releases API: ${mockReleasesId}`);
      spinner.succeed(`${prefix} Would create manifest: ${mockManifestId}`);
      
      return {
        manifestId: mockManifestId,
        releasesId: mockReleasesId,
        manifest
      };
    }

    const releasesDataRes = await turboClient.upload({
        data: JSON.stringify(releasesData, null, 2),
        dataItemOpts: {
            tags: [
                ...dataItemOptions.tags,
                {name: 'Content-Type', value: 'application/json'},
                {name: 'App-Version', value: version}
            ]
        }
    })

    
    // Add releases API to manifest
    manifest.paths['/releases/'] = {
      id: releasesDataRes.id
    };
    
    // Upload manifest
    const manifestTransaction = await turboClient.upload({
        data: JSON.stringify(manifest, null, 2),
        dataItemOpts: {
            tags: [
                ...dataItemOptions.tags,
                {name: 'Content-Type', value: 'application/x.arweave-manifest+json'},
                {name: 'App-Version', value: version}
            ]
        }
    })
    
    
    spinner.succeed(`Manifest created: ${manifestTransaction.id}`);
    return {
      manifestId: manifestTransaction.id,
      releasesId: releasesDataRes.id,
      manifest
    };
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    spinner.fail(`Failed to create manifest: ${errorMessage}`);
    throw error;
  }
}

/**
 * Main deployment function
 */
async function main(): Promise<void> {
  try {
    const title = CONFIG.dryRun ? '🎭 Harlequin CLI Arweave Deployment (DRY RUN)' : '🎭 Harlequin CLI Arweave Deployment';
    console.log(chalk.bold.blue(title));
    if (CONFIG.dryRun) {
      console.log(chalk.yellow('⚠️  DRY RUN MODE - No files will be uploaded to Arweave'));
    }
    console.log(chalk.gray('─'.repeat(50)));
    
    // Load wallet
    const wallet = await loadWallet();

    // create authenticated turbo, ario, and ANT clients
    const signer = new ArweaveSigner(wallet);
    const turboUploader = TurboFactory.authenticated({signer});
    
    let ario, ant;
    if (!CONFIG.dryRun) {
      ario = ARIO.init({
          signer,
          process: new AOProcess({
              processId: CONFIG.arns.registry,
              ao: connect({CU_URL: CONFIG.network.cuUrl, MODE: 'legacy'})
          })
      })

      const arnsRecord = await ario.getArNSRecord({name: CONFIG.arns.name});
      ant = ANT.init({
          signer,
          process: new AOProcess({
              processId: arnsRecord.processId,
              ao: connect({CU_URL: CONFIG.network.cuUrl, MODE: 'legacy'})
          })
      })
    }

    
    // Get version
    const version = getVersion();
    
    // Find binaries
    const binaries = findBinaries();
    
    if (binaries.length === 0) {
      throw new Error('No binary files found to upload');
    }


    
    // Upload install script if it exists
    const installScriptPath = join(__dirname, 'install_cli.sh');
    let installScript: TurboUploadDataItemResponse | null = null;
    
    if (existsSync(installScriptPath)) {
      if (CONFIG.dryRun) {
        installScript = {
          id: `mock_install_${Math.random().toString(36).substr(2, 43)}`,
          owner: 'mock',
          winc: '0',
          dataCaches: [],
          fastFinalityIndexes: []
        } as TurboUploadDataItemResponse;
      } else {
        installScript = await turboUploader.upload({
          data: readFileSync(installScriptPath),
          dataItemOpts: {
              tags: [
                  ...dataItemOptions.tags,
                  {name: "Content-Type", value: "application/x-shellscript"},
                  {name: "App-Version", value: version}
              ]
          }
        })
      }
    }
    
    // Upload all binaries
    console.log(chalk.blue(`\n📦 Uploading ${binaries.length} binary files...`));
    const uploads: {uploadResponse: TurboUploadDataItemResponse, name: string}[] = [];
    
    for (const binary of binaries) {
      const binaryName = basename(binary);
      const spinner = ora(`Processing ${binaryName}...`).start();
      
      if (CONFIG.dryRun) {
        const mockUpload = {
          id: `mock_binary_${Math.random().toString(36).substr(2, 43)}`,
          owner: 'mock',
          winc: '0',
          dataCaches: [],
          fastFinalityIndexes: []
        } as TurboUploadDataItemResponse;
        spinner.succeed(`[DRYRUN] Would compress and upload ${binaryName}`);
        uploads.push({uploadResponse: mockUpload, name: binaryName});
      } else {
        // Read and compress the binary
        const binaryData = readFileSync(binary);
        const compressedData = gzipSync(binaryData);
        const compressionRatio = ((1 - compressedData.length / binaryData.length) * 100).toFixed(1);
        
        spinner.text = `Uploading compressed ${binaryName} (${compressionRatio}% smaller)...`;
        
        const upload = await turboUploader.upload({
          data: compressedData,
          dataItemOpts: {
              tags: [
                  ...dataItemOptions.tags,
                  {name: "Content-Type", value: "application/gzip"},
                  {name: "Content-Encoding", value: "gzip"},
                  {name: "Original-Content-Type", value: "application/octet-stream"},
                  {name: "Original-Size", value: binaryData.length.toString()},
                  {name: "Compressed-Size", value: compressedData.length.toString()},
                  {name: "App-Version", value: version}
              ]
          }
        })
        
        spinner.succeed(`Uploaded ${binaryName} (compressed ${compressionRatio}%): ${upload.id}`);
        uploads.push({uploadResponse: upload, name: binaryName});
      }
    }
    
    // Create manifest
    const { manifestId, releasesId } = await createManifest(turboUploader, uploads, installScript);
    
    // Update ARNS
    if (CONFIG.dryRun) {
      console.log(chalk.yellow(`[DRYRUN] Would update ARNS undername: ${CONFIG.arns.undername}.${CONFIG.arns.name}`));
      console.log(chalk.yellow(`[DRYRUN] Would point to manifest: ${manifestId}`));
    } else {
      await ant!.setUndernameRecord({
          undername: CONFIG.arns.undername,
          transactionId: manifestId,
          ttlSeconds: 60
      }, {tags: dataItemOptions.tags})
    }
    
    // Summary
    const successMessage = CONFIG.dryRun ? '\n✅ Dry run completed successfully!' : '\n✅ Deployment completed successfully!';
    console.log(chalk.green(successMessage));
    if (CONFIG.dryRun) {
      console.log(chalk.yellow('ℹ️  This was a simulation - no files were actually uploaded'));
    }
    console.log(chalk.gray('─'.repeat(50)));
    console.log(chalk.blue(`📦 Version: ${version}`));
    console.log(chalk.blue(`📁 Files processed: ${uploads.length}`));
    console.log(chalk.blue(`🗂️  Manifest ID: ${manifestId}`));
    console.log(chalk.blue(`📊 Releases API: ${releasesId}`));
    if (!CONFIG.dryRun) {
      console.log(chalk.blue(`🌐 Preview: https://arweave.net/${manifestId}`));
    }
    console.log(chalk.blue(`🔗 ARNS URL: https://${CONFIG.arns.name}.arweave.dev`));
    
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    console.error(chalk.red('\n❌ Deployment failed:'));
    console.error(chalk.red(errorMessage));
    process.exit(1);
  }
}

// Run if called directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}

export { main };
