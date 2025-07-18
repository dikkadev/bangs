# Bangs

A lightweight and extensible search engine that leverages "bangs" to quickly redirect queries to your favorite services. Includes a web UI for browsing and searching available bangs. Customize your search experience by defining your own bangs and effortlessly expand its capabilities.

## Features

- **Web UI**: Browse, filter, and search available bangs through a modern web interface.
- **Customizable Bangs**: Define and manage your own bangs through a simple `bangs.yaml` configuration file.
- **Flexible Query Handling**: Perform searches using the `/bang?q=...` endpoint.
- **Flexible Default Search**: Specify a default as a URL, single bang, or multi-bang combination (e.g., `g+gh` opens both Google and GitHub).
- **Multi-Bang Support**: Use `+` to combine bangs (e.g., `!g+gh+so query` opens Google, GitHub, and StackOverflow).
- **Public Instances**: Accessible via [https://s.dikka.dev](https://s.dikka.dev) and [https://s.dikka.dev](https://s.dikka.dev) with HTTPS support (serving the web UI and handling bang redirects).
- **Extensible & Open**: Contributions to expand and refine bang definitions are highly encouraged.
- **Robust Testing**: Automated Go workflows ensure the correctness and reliability of bang URLs.
- **Categorized Bangs**: Organize bangs into categories for better management and user experience.

## Public Instances

Experience Bangs, including the web UI, without any setup by using our public instance:

- **Primary Instance**: [https://s.dikka.dev](https://s.dikka.dev)

The URL serves the web UI at the root, handles bang redirection via `/bang`, and supports HTTPS for secure connections.

## Installation

### Docker (Preferred)

The Docker image includes the pre-built web UI and the Go backend.

- **Docker Image**: Hosted on GitHub Container Registry (ghcr.io).
- **Usage**: 

  ```bash
  docker run -d -p 8080:8080 -v ./bangs.yaml:/app/bangs.yaml -e BANGS_BANGFILE=/app/bangs.yaml ghcr.io/dikkadev/bangs:latest
  ```

  Or use with Docker Compose:

  ```yaml
  services:
    bangs:
      image: ghcr.io/dikkadev/bangs:latest
      restart: unless-stopped
      pull_policy: always
      ports:
        - 8080:8080
      volumes:
        - ./bangs.yaml:/app/bangs.yaml # Mount your bangs config
      environment:
        - BANGS_BANGFILE=/app/bangs.yaml
        - BANGS_WATCH=true # Optional: Reload bangs.yaml on change
        # Add other BANGS_* environment variables as needed (see Command-Line Options below)
  ```

### Build from Source

Ensure you have [Go](https://golang.org/dl/) (version 1.21 or later) and [Bun](https://bun.sh/) installed.

1.  **Clone the Repository**

    ```bash
    git clone https://github.com/dikkadev/bangs.git
    cd bangs
    ```

2.  **Build the Frontend**

    ```bash
    cd frontend
    bun install
    bun run build
    cd .. # Go back to the root directory
    ```

3.  **Build the Go Application**

    ```bash
    go build -o bangs .
    ```

## Usage

1.  **Start the Application**:
    -   If built from source: `./bangs -b path/to/your/bangs.yaml`
    -   If using Docker: Use `docker run` or `docker compose up` as shown in Installation.

2.  **Access the Web UI**: Open your browser to `http://localhost:PORT` (e.g., `http://localhost:8080`). Here you can browse, search, and filter available bangs.

3.  **Perform Bang Redirects**: Use the `/bang` endpoint with the `q` query parameter.

    **Example (GitHub Search):**

    ```
    http://localhost:8080/bang?q=!gh dikkadev/bangs
    ```

    **Explanation:** The above URL uses the `!gh` bang to perform a GitHub search for the repository `dikkadev/bangs`.

    **Example (Default Search):**

    If a `default` URL is defined in `bangs.yaml`, queries to `/bang?q=...` without a recognized bang prefix will use the default search engine.

    ```
    http://localhost:8080/bang?q=OpenAI ChatGPT
    ```

    **Example (Double Hashtag Search):**

    Prefixing the query with `##` also forces the use of the default search engine.

    ```bash
    http://localhost:8080/bang?q=##OpenAI ChatGPT
    ```

4.  **API Endpoint:** The frontend uses the `/api/list` endpoint to fetch the available bangs data in JSON format.

## Command-Line Options & Environment Variables

The application can be configured via command-line flags or corresponding environment variables.

| Flag            | Env Variable            | Description                                      | Default         | Example                  |
|-----------------|-------------------------|--------------------------------------------------|-----------------|--------------------------|
| `--bangs`       | `BANGS_BANGFILE`        | Path to the YAML file containing bang definitions. | *(Required)*    | `-b bangs.yaml`          |
| `--port`        | `BANGS_PORT`            | Port on which the server will run.               | `8080`          | `-p 9090`                |
| `--watch`       | `BANGS_WATCH`           | Reload bangs file on change.                     | `false`         | `-w`                     |
| `--allow-no-bang`| `BANGS_ALLOW_NO_BANG`   | Allow `/bang` requests with no bang to be handled by default. | `false`         | `-a`                     |
| `--ignore-char` | `BANGS_IGNORE_CHAR`     | Start `/bang` query with this char to ignore bangs. | `.`             | `-i ~`                   |
| `--verbose`     | `BANGS_VERBOSE`         | Enable verbose debug logging.                    | `false`         | `-v`                     |
| `--help`        |                         | Show help message.                               | `false`         | `-h`                     |

*Note: Environment variables take precedence over default values, and command-line flags take precedence over environment variables.* 

## Configuration (`bangs.yaml`)

Bangs are defined in a `bangs.yaml` file. Each bang maps a unique *name* (used internally and in the UI) to its properties: `bang` characters, search `url` (with `{}` placeholder), `description`, and optional `category`.

The `default` key at the root specifies what happens for queries without a bang. This can be a URL, a single bang reference, or multiple bang references.

**Example `bangs.yaml`:**

```yaml
# URL default:
# default: 'https://www.google.com/search?q={}'

# Single bang reference as default:
# default: 'g'

# Multi-bang reference as default (opens multiple tabs):
default: 'g+gh'

# Other examples:
# default: 'ai+g'        # AI search + Google  
# default: 'g+gh+so'     # Google + GitHub + StackOverflow
# default: 'ddg'         # Just DuckDuckGo

# Bang Definitions (Key is the name)
GitHub:
  bang: 'gh'
  url: 'https://github.com/search?q={}'
  description: 'Search code repositories on GitHub'
  category: 'Development'

Google:
  bang: "g"
  url: "https://www.google.com/search?q={}"
  description: 'Popular global search engine by Google.'
  category: 'Search'

ChatGPT:
  bang: "ai"
  url: "https://chatgpt.com/?q={}"
  description: 'AI assistant for questions and tasks'
  category: 'AI'
```

### Aliases

Aliases allow you to create custom shortcuts for single bangs or multi-bang combinations. They are defined in the `aliases` section of your `bangs.yaml` file.

**Example with aliases:**

```yaml
# Aliases - Custom shortcuts for bangs
aliases:
  def: 'ai+g'        # !def -> AI search + Google
  search: 'g'        # !search -> Google only  
  shop: 'a+eb'       # !shop -> Amazon + eBay

# Regular bang definitions...
Google:
  bang: "g"
  url: "https://www.google.com/search?q={}"
  description: 'Popular global search engine'
  category: 'Search'

Amazon:
  bang: "a"
  url: "https://www.amazon.com/s?k={}"
  description: 'E-commerce platform'
  category: 'Shopping'
```

**Usage examples:**
- `!def machine learning` → Opens both ChatGPT and Google with "machine learning"
- `!search python` → Opens Google with "python"
- `!shop laptop` → Opens both Amazon and eBay with "laptop"

Aliases are displayed in the web UI with a special "Aliases" category and purple styling to distinguish them from regular bangs.

### Default Configuration Options

- **URL**: `default: 'https://www.google.com/search?q={}'` - URL with `{}` placeholder
- **Single Bang**: `default: 'g'` - Uses the Google bang for all default searches
- **Multi-Bang**: `default: 'g+gh'` - Opens both Google and GitHub in separate tabs
- **Complex Multi-Bang**: `default: 'ai+g+so'` - Opens AI, Google, and StackOverflow

## Setting Bangs as the Default Search Engine

You can configure your browser to use your running Bangs instance (local or public) as the default search engine. The exact method varies by browser, but generally involves setting the search URL to:

`http://YOUR_INSTANCE_URL/bang?q=%s`

Replace `YOUR_INSTANCE_URL` with the appropriate address (e.g., `localhost:8080`, `s.dikka.dev`).

### Chromium-Based Browsers

Search for adding a custom search engine in your browser's settings.

### Firefox

Requires an extension like "Add custom search engine".

### Pop-up / Redirect Settings

Multiple bangs (e.g., chaining multiple bangs in one query) open separate tabs or pop-up windows by design, which may be blocked by default. To ensure multibangs work as expected, allow pop-ups/redirects for your Bangs instance domain (e.g., `https://s.dikka.dev`) in your browser settings. Refer to your browser’s documentation for instructions on enabling pop-ups or redirects if needed.

## Advanced Usage

For details on advanced configurations and persistent setups, please refer to the [Advanced Usage](./ADVANCED.md) guide.


## License

[LICENSE](./LICENSE)

---

Happy Searching! 🚀
