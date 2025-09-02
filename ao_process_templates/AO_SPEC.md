# AOS Process Security Specification

## Overview

This specification defines the security-critical APIs and patterns for
implementing AOS (Arweave Operating System) processes across multiple
programming languages. The core security model revolves around message handling,
authority verification, assignment management, and state isolation.

## Core Architecture

### 1. Entry Point: Global Handle Function

```typescript
// Main entry point that all implementations must provide
function handle(messageJson: string, environmentJson: string): string
```

**Responsibilities:**

- Parse JSON input into structured message and environment objects
- Call `process.handle()` with parsed objects
- Return JSON-encoded result
- Handle top-level parsing errors

### 2. Process Handle Function

```typescript
function processHandle(msg: Message, env: Environment): ProcessResult
```

**Core Flow:**

1. Initialize AO environment (`ao.init()`)
2. Initialize process state
3. Normalize message (extract tags to root level)
4. Perform security checks
5. Handle boot functionality (if applicable)
6. Evaluate message handlers
7. Return structured result

## Security Model

### Message Trust Verification

```typescript
interface TrustValidation {
  // Rule 1: Message from owner is always trusted
  isOwnerMessage: boolean

  // Rule 2: Check against authorities list
  isFromAuthority: boolean

  // Rule 3: Verify signature authenticity
  isSignatureValid: boolean
}

function isTrusted(msg: Message): boolean {
  // Trust messages from signed owner
  if (msg.From === msg.Owner) return true

  // Trust messages from authorities
  for (const authority of ao.authorities) {
    if (msg.From === authority || msg.Owner === authority) {
      return true
    }
  }

  return false
}
```

### Assignment Security

```typescript
function isAssignment(msg: Message): boolean {
  return msg.Target !== ao.id
}

function isAssignable(msg: Message): boolean {
  if (ao.assignables.length === 0) return false

  for (const assignable of ao.assignables) {
    if (matchesSpec(msg, assignable.pattern)) {
      return true
    }
  }
  return false
}
```

**Security Rules:**

- Only trusted messages can perform assignments
- Assignments must match predefined assignable patterns
- Default deny: if no assignables configured, reject all assignments

## Data Structures

### Message Structure

```typescript
interface Message {
  // Core identifiers
  Id: string
  From: string
  Owner: string
  Target: string

  // Content
  Data: any
  Tags: Tag[]
  TagArray: Tag[] // Original tags array

  // Metadata
  Timestamp: number
  'Block-Height': string
  Reference?: string
  'X-Reference'?: string
  'Reply-To'?: string
  'X-Origin'?: string

  // Methods (added by process)
  reply?: (replyMsg: Partial<Message>) => void
  forward?: (target: string, forwardMsg?: Partial<Message>) => void
}

interface Tag {
  name: string
  value: string
}
```

### Environment Structure

```typescript
interface Environment {
  Process: {
    Id: string
    Tags: Tag[]
    TagArray: Tag[] // Normalized version
  }
}
```

### AO Object Structure

```typescript
interface AO {
  // Core properties
  _version: string
  id: string
  _module: string
  authorities: string[]
  reference: number

  // Outbox for managing outputs
  outbox: {
    Output: any
    Messages: Message[]
    Spawns: SpawnRequest[]
    Assignments: Assignment[]
    Error?: string
  }

  // Security lists
  nonExtractableTags: string[]
  nonForwardableTags: string[]
  assignables: AssignableSpec[]

  // Core functions
  init: (env: Environment) => void
  send: (msg: Partial<Message>) => MessageHandle
  spawn: (module: string, msg: Partial<Message>) => SpawnHandle
  assign: (assignment: Assignment) => void
  isTrusted: (msg: Message) => boolean
  clearOutbox: () => void
  result: (result: any) => ProcessResult
}
```

## Security Implementation Requirements

### 1. Message Validation

```typescript
function validateMessage(msg: Message, env: Environment): SecurityResult {
  // Check 1: Trust validation
  if (msg.From !== msg.Owner && !ao.isTrusted(msg)) {
    return {
      trusted: false,
      error: 'Message is not trusted by this process!',
      action: 'send_error_to_sender',
    }
  }

  // Check 2: Assignment validation
  if (ao.isAssignment(msg) && !ao.isAssignable(msg)) {
    return {
      trusted: false,
      error: 'Assignment is not trusted by this process!',
      action: 'send_error_to_sender',
    }
  }

  return { trusted: true }
}
```

### 2. State Isolation

```typescript
function initializeState(msg: Message, env: Environment): void {
  // Initialize global state only once
  if (!Seeded) {
    // Deterministic seeding based on message properties
    const seed = generateSeed(msg['Block-Height'], msg.Owner, msg.Module, msg.Id)
    seedRandomGenerator(seed)
    Seeded = true
  }

  // Initialize process-specific state
  Errors = Errors || []
  Inbox = Inbox || []

  // Set ownership from environment or message
  if (!Owner) {
    Owner = findOwnerFromProcess(env) || msg.From
  }

  // Set process name
  if (!Name) {
    Name = findNameFromProcess(env) || 'aos'
  }
}
```

### 3. Outbox Management

```typescript
function clearOutbox(): void {
  ao.outbox = {
    Output: {},
    Messages: [],
    Spawns: [],
    Assignments: [],
  }
}

function addToOutbox(type: 'Messages' | 'Spawns' | 'Assignments', item: any): void {
  ao.outbox[type].push(item)
}
```

## Boot Functionality

### Boot Handler Implementation

```typescript
function handleBoot(msg: Message): void {
  // Only process boot messages from owner
  if (msg.Tags.Type !== 'Process' || Owner !== msg.From) {
    return
  }

  // Handle state initialization
  if (msg.Data && typeof msg.Data === 'string') {
    try {
      initializeANTState(msg.Data)
    } catch (error) {
      sendBootError(msg, error)
    }
  }

  // Send success notifications
  sendBootNotifications(msg)
}
```

## Error Handling

### Error Response Structure

```typescript
interface ErrorResult {
  Error: string
  Messages?: Message[]
  Spawns?: SpawnRequest[]
  Assignments?: Assignment[]
}

interface SuccessResult {
  Output: any
  Messages: Message[]
  Spawns: SpawnRequest[]
  Assignments: Assignment[]
}

type ProcessResult = ErrorResult | SuccessResult
```

### Error Handling Pattern

```typescript
function handleWithErrorCapture(handler: Function, msg: Message): ProcessResult {
  try {
    const result = handler(msg)
    return ao.result(result)
  } catch (error) {
    // Log error for debugging
    console.error('Handler error:', error)

    // Return error result
    return {
      Error: error.message,
      Messages: [],
      Spawns: [],
      Assignments: [],
    }
  }
}
```

## Implementation Checklist

### Core Requirements

- [ ] JSON parsing/encoding for message and environment
- [ ] Message trust validation against authorities
- [ ] Assignment security checks
- [ ] State isolation and initialization
- [ ] Outbox management (Messages, Spawns, Assignments)
- [ ] Boot handler for process initialization
- [ ] Error handling and result formatting
- [ ] Tag normalization (TagArray → Tags object)

### Security Requirements

- [ ] Authority list verification
- [ ] Owner validation
- [ ] Assignment pattern matching
- [ ] Non-extractable tag protection
- [ ] Non-forwardable tag filtering
- [ ] Deterministic random seeding
- [ ] Process ID validation

### Message Flow Requirements

- [ ] Reply message composition
- [ ] Forward message sanitization
- [ ] Reference tracking
- [ ] Inbox management with overflow protection
- [ ] Output formatting with prompts

## Language-Specific Considerations

### JavaScript/TypeScript

- Use proper JSON parsing with error handling
- Implement deep cloning for message sanitization
- Handle async operations in handlers appropriately

### Rust

- Use serde for JSON serialization/deserialization
- Implement proper error handling with Result types
- Ensure memory safety in message passing

### Go

- Use encoding/json for message parsing
- Implement proper error handling with error interface
- Ensure goroutine safety if using concurrency

### Python

- Use json module for parsing
- Implement proper exception handling
- Consider using dataclasses for message structures

## Testing Requirements

### Security Tests

- [ ] Untrusted message rejection
- [ ] Invalid assignment blocking
- [ ] Authority verification
- [ ] Owner-only operations

### Functional Tests

- [ ] Message routing and replies
- [ ] Boot process initialization
- [ ] Error handling paths
- [ ] Outbox state management

### Integration Tests

- [ ] End-to-end message processing
- [ ] Multi-message state persistence
- [ ] Cross-process communication patterns

## Security Considerations

1. **Input Validation**: All JSON inputs must be validated before processing
2. **State Isolation**: Each process must maintain isolated state
3. **Authority Verification**: Never trust messages without proper authority
   checks
4. **Assignment Control**: Strict pattern matching for assignment operations
5. **Error Information**: Avoid leaking sensitive information in error messages
6. **Deterministic Behavior**: Random seeding must be deterministic for replay
   consistency

This specification ensures consistent, secure implementation of AOS processes
across different programming languages while maintaining the core security model
and message handling patterns.
