import React, { ReactNode } from 'react'
import { LanguageSelector } from '../language-selector'
import { Button } from '../ui/button'
import { ThemeToggle } from '../theme-toggle'
import { useTranslation } from 'react-i18next'
import { Github } from 'lucide-react'

interface IProps {
  leftNode?: ReactNode
}
export function Header(props: IProps) {
  const { t } = useTranslation()

  return (
    <div className="bg-background/80 fixed left-0 top-0 z-50 flex w-full items-center justify-between border-b px-4 py-4 backdrop-blur-sm md:px-12">
      <a href="/" className="text-xs font-semibold text-foreground transition-colors hover:text-primary md:text-base">
        Vite React TS Tailwind Starter
      </a>
      <div className="flex items-center gap-4">
        <LanguageSelector />
        <ThemeToggle />
        <Button size={'icon'} asChild className="rounded-full">
          <a href="https://github.com/Quilljou/vite-react-ts-tailwind-starter" target="_blank" rel="noreferrer">
            <Github />
          </a>
        </Button>
      </div>
    </div>
  )
}
