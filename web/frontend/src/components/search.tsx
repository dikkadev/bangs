"use client"

import type React from "react"

// Remove unused useState import
// import { useState } from "react"
import { SearchIcon, ArrowRightIcon } from "lucide-react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

// Define props type
interface SearchProps {
  query: string;
  onQueryChange: (query: string) => void;
}

export function Search({ query, onQueryChange }: SearchProps) {
  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()

    if (!query.trim()) return

    // Check if query has a bang prefix
    const hasBang = query.startsWith("!")

    let targetUrl = "";

    // Construct the target URL
    if (hasBang) {
      // Extract the bang and the search term
      const parts = query.split(" ")
      const bang = parts[0]
      const searchTerm = parts.slice(1).join(" ")

      // Target the public instance with the bang
      targetUrl = `https://s.dikka.dev/?q=${encodeURIComponent(bang + " " + searchTerm)}`
    } else {
      // Use default search
      targetUrl = `https://s.dikka.dev/?q=${encodeURIComponent(query)}`
    }

    // Open in new tab
    window.open(targetUrl, "_blank", "noopener,noreferrer");
  }

  return (
    <div className="relative">
      <form onSubmit={handleSearch} className="relative">
        <div className="relative flex items-center">
          <div className="absolute left-0 top-0 bottom-0 w-12 flex items-center justify-center z-10 pointer-events-none">
            <SearchIcon className="text-pink-500 h-5 w-5" />
          </div>
          <Input
            type="text"
            placeholder="Search with or without a bang prefix..."
            value={query}
            onChange={(e) => onQueryChange(e.target.value)}
            className="pl-12 pr-12 h-12 bg-black border-white/20 text-white placeholder:text-gray-500 focus:border-pink-500 focus:ring-1 focus:ring-pink-500 w-full"
            aria-label="Search query"
          />
          <Button
            type="submit"
            size="icon"
            className="absolute right-2 top-1/2 transform -translate-y-1/2 h-8 w-8 bg-pink-500 hover:bg-pink-600 text-white"
            aria-label="Submit search"
          >
            <ArrowRightIcon className="h-4 w-4" />
          </Button>
        </div>
      </form>
      <div className="mt-4 text-sm text-gray-400 flex items-center">
        <div className="w-4 h-4 rounded-full bg-pink-500/20 border border-pink-500/50 mr-2 flex items-center justify-center">
          <span className="text-xs text-pink-500">i</span>
        </div>
        <span>
          Works with bangs (<span className="text-pink-500 font-mono">!gh query</span>) or without (
          <span className="text-pink-500 font-mono">query</span>)
        </span>
      </div>
    </div>
  )
}
