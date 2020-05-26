## Path of Exile Arbitrage

CLI tool to detect arbitrage opportunities by taking advantage of inefficient
bid-ask spreads in the [Bulk Item Exchange](https://www.pathofexile.com/trade/exchange).
This tool relies on the exchange API which is not officially supported by GGG.

## Installation

A pre-built Windows 64-bit executable can be found under releases.

You can also build an executable directly by installing Go, cloning the repo
and running `go build poe-arbitrage`. Additional documentation can be found
[here](https://golang.org/cmd/go/#hdr-Compile_packages_and_dependencies).

## Feature

- [x] Fetch bulk trades from PoE Bulk Exchange API
- [x] Exclude AFK users by default
- [x] Ignore users
- [x] Favorite users
- [ ] Print whispers for profitable arbitrage opportunities

## Usage

```sh
# List all supported bulk items
poe-arbitrage list

# List bulk items with name containing "orb of" (case insensitive)
poe-arbitrage list --name "orb of"

# Check for opportunities when trading Chaos Orbs or Exalt Orbs (at least 2 items)
poe-arbitrage trade chaos exa

# Check for opportunities in the current hardcore league
poe-arbitrage trade chaos exa gcp

# Check for opportunities with 100 Chaos, 0 Exalts and 20 GCPs
poe-arbitrage trade chaos exa gcp --capital chaos=100,gcp=20

# Configure the CLI behavior via CLI
# The config is stored as JSON locally and can be manually edited.
poe-arbitrage configure --league Standard
poe-arbitrage configure --hardcore false
poe-arbitrage configure --exclude-afk true
poe-arbitrage configure --ignore-player ABC
poe-arbitrage configure --favorite-player XYZ
poe-arbitrage configure --set-item "golden-oil,Golden Oil,10"
```

Given `N` items, CLI makes `2 * N!/(N-2)!` number of API calls to determine
possible trading opportunities. `N!/(N-2)!` is the number of trading pairs
(order matters). Often it is more profitable to trade to an intermediate
item rather than trading two items directly.

Some suggestions to cut down the number of API calls is selecting popular
items with a high stack size and high innate value. Since users may be
unresponsive, its important to choose items that you do not mind holding
for extended periods of time.

## Open Questions

- How will the stack size affect trades?
  - Example: Exalt Orbs for Chaos Orbs where one trade is not enough.
  - Current limiting trades to `5 * 12 * stack size`
- What are the API rate limits for POE exchange API?
- Profitability of flipping inefficiently priced items.
  - Price rare items via ML.
  - Flipping via vendor recipes (e.g. quality gems or higher tier essences)
