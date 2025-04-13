import { useState } from "react"
import { Search } from "./components/search"
import { BangsList } from "./components/bangs-list"
import { GithubIcon, ExternalLink } from "lucide-react"

export default function Home() {
  const [searchQuery, setSearchQuery] = useState("");

  return (
    <div className="min-h-screen text-white overflow-hidden bg-black">
      {/* Diagonal accent line */}
      <div className="fixed top-0 left-0 w-full h-1 bg-gradient-to-r from-purple-500 via-pink-500 to-orange-500 transform -rotate-0 origin-top-left z-10"></div>

      {/* Background grid */}
      <div
        className="fixed inset-0 opacity-20 pointer-events-none z-5"
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fill, minmax(50px, 1fr))",
          gridTemplateRows: "repeat(auto-fill, minmax(50px, 1fr))",
        }}
      >
        {Array.from({ length: 4000 }).map((_, i) => (
          <div key={i} className="border-[0.2px] border-white/20 aspect-square"></div>
        ))}
      </div>


      {/* Main content */}
      <div className="relative z-20">
        {/* Header - positioned on the left side, hidden on small screens */}
        <header className="fixed top-0 left-0 h-full w-64 flex-col justify-center items-center border-r border-white/10 bg-black backdrop-blur-sm hidden md:flex">
          <div className="p-8 flex flex-col items-center">
            <h1 className="text-4xl font-bold mb-4 tracking-tight">
              <span className="text-pink-500">!</span>bangs
            </h1>
            <div className="w-12 h-1 bg-gradient-to-r from-purple-500 to-orange-500 mb-6"></div>
            <p className="text-gray-400 text-sm text-center mb-8">Redirect searches to your favorite services</p>
            <div className="flex gap-4 mt-4">
              <a
                href="https://github.com/Sett17/bangs"
                target="_blank"
                rel="noopener noreferrer"
                className="text-gray-400 hover:text-white transition-colors"
                aria-label="GitHub Repository"
              >
                <GithubIcon className="w-5 h-5" />
              </a>
              <a
                href="https://bang.dikka.dev"
                target="_blank"
                rel="noopener noreferrer"
                className="text-gray-400 hover:text-white transition-colors"
                aria-label="Public Instance"
              >
                <ExternalLink className="w-5 h-5" />
              </a>
            </div>
          </div>
        </header>

        {/* Main content area - offset on medium screens and up */}
        <main className="p-8 min-h-screen md:ml-64 z-9">
          {/* Search section - positioned at the top */}
          <section className="mb-16 pt-8">
            <div className="max-w-3xl">
              <h2 className="text-2xl font-bold mb-6 text-white flex items-center">
                <span className="text-pink-500 mr-2">01</span>
                Search
              </h2>
              <Search query={searchQuery} onQueryChange={setSearchQuery} />
            </div>
          </section>

          {/* Bangs directory - takes up the rest of the space */}
          <section>
            <h2 className="text-2xl font-bold mb-6 text-white flex items-center">
              <span className="text-pink-500 mr-2">02</span>
              Directory
            </h2>
            <BangsList mainQuery={searchQuery} />
          </section>
        </main>
      </div>
    </div>
  )
}
