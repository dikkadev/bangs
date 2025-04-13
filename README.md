# Bangs

A lightweight and extensible search engine that leverages "bangs" to quickly redirect queries to your favorite services. Customize your search experience by defining your own bangs and effortlessly expand its capabilities.

## Features

- **Customizable Bangs**: Define and manage your own bangs through a simple `bangs.yaml` configuration file.
- **Flexible Query Handling**: Perform searches using query parameters or clean URL paths.
- **Default Search Engine**: Specify a default search URL for queries without a bang (only works when using `q` query parameter).
- **Public Instances**: Accessible via [https://bang.dikka.dev](https://bang.dikka.dev) and [https://s.dikka.dev](https://s.dikka.dev) with HTTPS support.
- **Extensible & Open**: Contributions to expand and refine bang definitions are highly encouraged.
- **Robust Testing**: Automated Go workflows ensure the correctness and reliability of bang URLs.
- **Categorized Bangs**: Organize bangs into categories for better management and user experience.

## Public Instances

Experience Bangs without any setup by using our public instances:

- **Primary Instance**: [https://bang.dikka.dev](https://bang.dikka.dev)
- **Alternative Instance**: [https://s.dikka.dev](https://s.dikka.dev)

Both URLs point to the same backend and support HTTPS for secure connections.

## Installation

### Docker (Preferred)

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
        - ./bangs.yaml:/app/bangs.yaml
      environment:
        - BANGS_BANGFILE=/app/bangs.yaml
        - BANGS_WATCH=true
  ```

### Build from Source

Ensure you have [Go](https://golang.org/dl/) installed (version 1.18 or later).

1. **Clone the Repository**

    ```bash
    git clone https://github.com/dikkadev/bangs.git
    cd bangs
    ```

2. **Build the Application**

    ```bash
    go build -o bangs ./cmd
    ```

## Usage

### Usage Scenarios

#### 1. Query-Based Search

Perform a search by specifying the bang and query as URL parameters using the public instance.

**Example:**

```bash
https://s.dikka.dev/?q=!gh dikkadev/bangs
```

**Explanation:**  
The above URL uses the `!gh` bang to perform a GitHub search for the repository `dikkadev/bangs`.

With query-based searches, you also have the option to omit the bang and use the default search engine defined in `bangs.yaml`. This is useful when you want to use bangs as the default search engine in your browser, so you do not need to specify a bang every time.

#### 2. Path-Based Search

Perform a search by embedding the bang and query directly into the URL path using the public instance.

**Example:**

```bash
https://s.dikka.dev/!gh/dikkadev/bangs
```

**Explanation:**  
This URL achieves the same search as the query-based example but uses the path to specify the bang and query.

#### 3. Default Search

Perform a search without specifying a bang, which uses the default search engine.

**Example:**

```bash
https://s.dikka.dev/?q=OpenAI ChatGPT
```

**Explanation:**  
The above URL uses the default search engine defined in `bangs.yaml` to search for "OpenAI ChatGPT".

#### 4. Double Hashtag Search

Perform a search by typing a double hashtag `##` at the start of the query, which uses the default search engine.

**Example:**

```bash
https://s.dikka.dev/?q=##OpenAI ChatGPT
```

**Explanation:**  
The above URL uses the default search engine defined in `bangs.yaml` to search for "OpenAI ChatGPT" by stripping the `##` from the query.

### Command-Line Options

| Option          | Short | Description                                     | Default         | Example                  |
|-----------------|-------|-------------------------------------------------|-----------------|--------------------------|
| `--bangs`       | `-b`  | Path to the YAML file containing bang definitions. | `bangs.yaml`    | `-b path/to/your/bangs.yaml` |
| `--port`        | `-p`  | Port on which the server will run.              | `8080`          | `-p 9090`                |
| `--verbose`     | `-v`  | Enable verbose debug logging.                   | `false`         | `-v`                     |
| `--help`        | `-h`  | Show help message.                              | `false`         | `-h`                     |

## Configuration

Bangs are defined in a `bangs.yaml` file. Each bang maps a unique identifier to a search URL containing a placeholder `{}` where the query will be inserted. Additionally, a `default` key can be specified to handle searches without a bang, and a `category` field can be used to organize bangs into groups.

**Example `bangs.yaml`:**

```yaml
default: 'https://www.google.com/search?q={}'

# Other search engines:
# default: 'https://www.bing.com/search?q={}'
# default: 'https://duckduckgo.com/?q={}'

GitHub:
  bang: 'gh'
  url: 'https://github.com/search?q={}'
  description: 'Search code repositories on GitHub'
  category: 'Development'

g:
  bang: "g"
  url: "https://www.google.com/search?q={}"
  description: 'Popular global search engine by Google.'
  category: 'Search'
```

## Setting Bangs as the Default Search Engine

This capability varies based on your browser.

### Chromium-Based Browsers

[s.dikka.dev/g/chrome set default search engine](https://s.dikka.dev/g/chrome%20set%20default%20search%20engine)

### Firefox

[s.dikka.dev/g/firefox set default search engine](https://s.dikka.dev/g/firefox%20set%20default%20search%20engine)

## Advanced Usage

Enhance your Bangs experience with advanced configurations and persistent setups. For detailed instructions, please refer to the [Advanced Usage](./ADVANCED.md) guide.

## Contributing

Contributions are highly welcome! If you have new bangs to add or improvements to the existing ones, feel free to submit a pull request to the `bangs.yaml` file.

### How to Contribute

- **Branch Naming**: Please name your feature branches in the format `bang/{chars}`, where `{chars}` represents the bang characters (e.g., `bang/y`, `bang/gh`).
- **Adding Bangs**: Ensure your new bang is placed in the appropriate section within `bangs.yaml`.
- **Non-Search Bangs**: If your contribution is not a search-related bang, please explain the use case clearly in your pull request description.
- **Follow Templates**: There are issue and PR templates in place to guide your contributions. Please follow them.

## Testing

The project includes a comprehensive Go test workflow that not only checks the functionality of the application but also validates the correctness of all bang URLs defined in `bangs.yaml`.

### Running Tests Locally

```bash
go test ./...
```

### Continuous Integration

Every pull request triggers the Go test workflow to ensure that new contributions do not break existing functionality and that all bang URLs remain valid.

## Docker

### Installation and Running with Docker

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
        - ./bangs.yaml:/app/bangs.yaml
      environment:
        - BANGS_BANGFILE=/app/bangs.yaml
        - BANGS_WATCH=true
  ```

## License

[LICENSE](./LICENSE)

---

Happy Searching! 🚀
