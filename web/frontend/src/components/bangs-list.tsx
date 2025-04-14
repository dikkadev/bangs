"use client"

import { useState, useEffect } from "react"
import { Input } from "@/components/ui/input"
import { Search, Copy, Check, ExternalLink } from "lucide-react"
import { Button } from "@/components/ui/button"
import { toast } from "@/components/ui/use-toast"
import { Toaster } from "@/components/ui/toaster"

// Type definition for the backend Entry structure + added name field
interface BangEntry {
  name: string // Added name field
  bang: string
  description: string
  url: string
  category: string
}

// Define the expected API response type (map where key is name)
type BangsApiResponse = Record<string, BangEntry>

// Define props type
interface BangsListProps {
  mainQuery: string;
}

// Custom category sort order (most useful first for technical users)
const categorySortOrder: string[] = [
  "Development",
  "Search",
  "Reference",
  "Media",
  "Shopping",
  "Social",
  // Add other categories here in desired order
];

// Helper function to generate the final URL
const generateFinalUrl = (baseUrl: string, query: string): string => {
  if (!query) {
    // If query is empty, just replace placeholder with literal {}
    return baseUrl.replace("{}", "{}");
  }
  // Otherwise, replace with encoded query
  return baseUrl.replace("{}", encodeURIComponent(query));
};

// Helper function to add period if needed
const ensurePeriod = (text: string): string => {
  if (!text) return "";
  const lastChar = text.trim().slice(-1);
  if ([".", "?", "!"].includes(lastChar)) {
    return text;
  }
  return text.trim() + ".";
};

export function BangsList({ mainQuery }: BangsListProps) { // Destructure props
  // const [allBangs, setAllBangs] = useState<BangsApiResponse>({}) // Store API response
  const [bangsList, setBangsList] = useState<BangEntry[]>([]) // Store processed list for rendering
  const [searchTerm, setSearchTerm] = useState("")
  const [copiedIndex, setCopiedIndex] = useState<string | null>(null) // Use bang name as key
  const [activeCategory, setActiveCategory] = useState<string | null>(null)
  const [isGridView, setIsGridView] = useState(true)
  const [categories, setCategories] = useState<string[]>([])

  // Fetch bangs from the API
  useEffect(() => {
    const fetchBangs = async () => {
      try {
        const response = await fetch("/bang/list")
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }
        const data: BangsApiResponse = await response.json()
        // Remove setAllBangs call
        // setAllBangs(data)

        // Process the map into an array
        let processedList = Object.entries(data).map(([name, entry]) => ({ ...entry, name })) // Add name field
        
        // Sort the list with custom category order
        processedList.sort((a, b) => {
          const categoryA = a.category || "";
          const categoryB = b.category || "";

          const indexA = categorySortOrder.indexOf(categoryA);
          const indexB = categorySortOrder.indexOf(categoryB);

          // Handle categories present in the custom order
          if (indexA !== -1 && indexB !== -1) {
            if (indexA < indexB) return -1;
            if (indexA > indexB) return 1;
          } else if (indexA !== -1) {
            return -1; // A is in order, B is not -> A comes first
          } else if (indexB !== -1) {
            return 1; // B is in order, A is not -> B comes first
          } else {
            // Neither A nor B are in the custom order, sort alphabetically (empty last)
            const catSortA = categoryA || "zzzz";
            const catSortB = categoryB || "zzzz";
            if (catSortA < catSortB) return -1;
            if (catSortA > catSortB) return 1;
          }

          // If categories are effectively equal (same custom index or both not found & alphabetically same),
          // sort by name alphabetically
          if (a.name < b.name) return -1;
          if (a.name > b.name) return 1;
          return 0;
        });

        setBangsList(processedList)

        // Update category filter buttons based on the *custom* order + remaining
        const uniqueCategories = Array.from(new Set(processedList.map((bang) => bang.category).filter(Boolean)))
        uniqueCategories.sort((a, b) => {
          const indexA = categorySortOrder.indexOf(a);
          const indexB = categorySortOrder.indexOf(b);
          if (indexA !== -1 && indexB !== -1) return indexA - indexB;
          if (indexA !== -1) return -1;
          if (indexB !== -1) return 1;
          return a.localeCompare(b); // Sort remaining alphabetically
        });
        setCategories(uniqueCategories)

      } catch (error) {
        console.error("Failed to fetch bangs:", error)
        toast({
          variant: "destructive",
          title: "Error fetching bangs",
          description: "Could not load bang data from the server.",
        })
      }
    }
    fetchBangs()
  }, [])

  const handleCopy = (url: string, name: string) => { // Use name instead of index
    // Use mainQuery prop for final URL generation
    const finalUrl = generateFinalUrl(url, mainQuery);

    navigator.clipboard.writeText(finalUrl)
    setCopiedIndex(name) // Use name as key

    toast({
      title: "URL copied!",
      // Update description to reflect the actual query used
      description: mainQuery ? `Copied with "${mainQuery}" as the search term` : "Copied base URL (no query)",
      duration: 3000,
    })

    setTimeout(() => {
      setCopiedIndex(null)
    }, 2000)
  }

  // Function to handle opening link in new tab
  const handleOpenLink = (url: string) => {
    const finalUrl = generateFinalUrl(url, mainQuery);
    window.open(finalUrl, "_blank", "noopener,noreferrer");
  }

  // Filter bangs based on search term and active category
  const filteredBangs = bangsList.filter( // Use bangsList state
    (bang) =>
      (activeCategory === null || bang.category === activeCategory) &&
      (bang.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        bang.bang.toLowerCase().includes(searchTerm.toLowerCase()) ||
        bang.description.toLowerCase().includes(searchTerm.toLowerCase())),
  )

  return (
    <div>
      <div className="mb-8 flex flex-wrap gap-4 items-center">
        <div className="relative flex-grow max-w-md">
          <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-pink-500">
            <Search className="h-4 w-4" />
          </div>
          <Input
            type="text"
            placeholder="Filter bangs..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10 bg-black border-white/20 text-white placeholder:text-gray-500 focus:border-pink-500 focus:ring-1 focus:ring-pink-500"
            aria-label="Filter bangs"
          />
        </div>

        <div className="flex gap-2">
          <Button
            variant={isGridView ? "default" : "outline"}
            size="sm"
            onClick={() => setIsGridView(true)}
            className={`${isGridView ? "bg-pink-500 hover:bg-pink-600" : "border-white/20 text-white bg-black hover:bg-white/10"}`}
            aria-label="Grid view"
          >
            Grid
          </Button>
          <Button
            variant={!isGridView ? "default" : "outline"}
            size="sm"
            onClick={() => setIsGridView(false)}
            className={`${!isGridView ? "bg-pink-500 hover:bg-pink-600" : "border-white/20 text-white bg-black hover:bg-white/10"}`}
            aria-label="List view"
          >
            List
          </Button>
        </div>
      </div>

      <div className="mb-6 flex flex-wrap gap-2">
        <Button
          variant={activeCategory === null ? "default" : "outline"}
          size="sm"
          onClick={() => setActiveCategory(null)}
          className={`${activeCategory === null ? "bg-pink-500 hover:bg-pink-600" : "border-white/20 text-white bg-black hover:bg-white/10"}`}
          disabled={!categories.length} // Disable if no categories loaded
        >
          All
        </Button>
        {categories.map((category) => (
          <Button
            key={category}
            variant={activeCategory === category ? "default" : "outline"}
            size="sm"
            onClick={() => setActiveCategory(category)}
            className={`${activeCategory === category ? "bg-pink-500 hover:bg-pink-600" : "border-white/20 text-white bg-black hover:bg-white/10"}`}
          >
            {category}
          </Button>
        ))}
      </div>

      {isGridView ? (
        <div className="grid grid-cols-2 gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"> {/* Default to 2 cols */}
          {filteredBangs.map((bang) => { // Add braces for variable declaration
            const displayUrl = generateFinalUrl(bang.url, mainQuery); // Generate URL for display
            return (
              <div
                key={bang.name} 
                className="group relative bg-black border border-white/10 hover:border-pink-500/50 transition-colors p-4 overflow-hidden"
              >
                <div className="absolute top-0 left-0 w-1 h-full bg-pink-500 opacity-0 group-hover:opacity-100 transition-opacity"></div>
                <div className="flex justify-between items-start mb-2">
                  <h3 className="font-bold text-white group-hover:text-pink-500 transition-colors">{bang.name}</h3>
                  <span className="text-pink-500 font-mono text-sm">{bang.bang}</span>
                </div>
                <p className="text-sm text-gray-400 mb-3">
                  {/* Category first */}
                  {bang.category && (
                    <span className="text-pink-500 text-sm mr-2">{bang.category}</span>
                  )}
                  {ensurePeriod(bang.description)}
                </p>
                <div className="flex items-center justify-between">
                  {/* Display generated URL */}
                  <div className="text-xs text-gray-500 font-mono break-all mr-2">{displayUrl}</div>
                  {/* Button Group */}
                  <div className="flex space-x-1 shrink-0">
                    {/* Open Link Button */}
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-8 w-8 border-white/10 hover:bg-pink-500 hover:text-white hover:border-pink-500 cursor-pointer"
                      onClick={() => handleOpenLink(bang.url)}
                      aria-label="Open link in new tab"
                    >
                      <ExternalLink className="h-4 w-4" />
                    </Button>
                    {/* Copy Button */}
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-8 w-8 border-white/10 hover:bg-pink-500 hover:text-white hover:border-pink-500 cursor-pointer"
                      onClick={() => handleCopy(bang.url, bang.name)}
                      aria-label="Copy URL with current search query"
                    >
                      {copiedIndex === bang.name ? <Check className="h-4 w-4 text-white" /> : <Copy className="h-4 w-4" />}
                    </Button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      ) : (
        <div className="border border-white/10">
          {filteredBangs.map((bang) => { // Add braces for variable declaration
            const displayUrl = generateFinalUrl(bang.url, mainQuery); // Generate URL for display
            return (
              <div
                key={bang.name}
                className="group flex items-center border-b border-white/10 last:border-b-0 bg-black transition-colors"
              >
                <div className="w-1 h-full bg-pink-500 opacity-0 group-hover:opacity-100 transition-opacity"></div>
                <div className="p-4 flex-grow">
                  <div className="flex items-center gap-2 mb-1">
                    <h3 className="font-bold text-white group-hover:text-pink-500 transition-colors">{bang.name}</h3>
                    <span className="text-pink-500 font-mono text-sm">{bang.bang}</span>
                  </div>
                  <p className="text-sm text-gray-400">
                    {/* Category first */}
                    {bang.category && (
                      <span className="text-pink-500 text-sm mr-2">{bang.category}</span>
                    )}
                    {ensurePeriod(bang.description)}
                  </p>
                  {/* Display generated URL */}
                  <div className="text-xs text-gray-500 font-mono mt-1">{displayUrl}</div>
                </div>
                <div className="p-4 shrink-0">
                  {/* Button Group */}
                  <div className="flex space-x-1">
                    {/* Open Link Button */}
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-8 w-8 border-white/10 hover:bg-pink-500 hover:text-white hover:border-pink-500 cursor-pointer"
                      onClick={() => handleOpenLink(bang.url)}
                      aria-label="Open link in new tab"
                    >
                      <ExternalLink className="h-4 w-4" />
                    </Button>
                    {/* Copy Button */}
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-8 w-8 border-white/10 hover:bg-pink-500 hover:text-white hover:border-pink-500 cursor-pointer"
                      onClick={() => handleCopy(bang.url, bang.name)}
                      aria-label="Copy URL with current search query"
                    >
                      {copiedIndex === bang.name ? <Check className="h-4 w-4 text-white" /> : <Copy className="h-4 w-4" />}
                    </Button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {filteredBangs.length === 0 && (
        <div className="text-center py-12 text-gray-400 border border-white/10 bg-black/30">
          No bangs found matching your search.
        </div>
      )}

      <Toaster />
    </div>
  )
}
