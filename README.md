## Path of Exile Arbitrage

CLI tool to detect arbitrage opportunities by taking advantage of inefficient
bid-ask spreads in the [Bulk Item Exchange](https://www.pathofexile.com/trade).
This tool relies on the unofficial POE exchange API.

Given `N` currencies, `poe-arbitrage` makes `2 * (N choose 2)` number of API calls
to determine possible trading opportunities. `N choose 2` is the number of unique
currency pairs. Often it is more profitable to trade to an intermediate
currency rather than trading two currencies directly.

Some suggestions to cut down the number of API calls is selecting popular
currencies with a high stack size and high innate value. Since users may be
unresponsive, its important to choose currencies that you do not mind holding
for extended periods of time.

## Feature

- Detect opportunities with a minimum of `N` online users.
  - Decreases potential transaction costs and opportunity costs of holding
    non-liquid currencies.
  - Optional AFK filter
- Detect opportunities with a minimum of `N` profit (dependent on trading pair)
- Set initial capital via JSON/CLI
- Limit trade pairs via JSON/CLI
- Limit per trade volume via JSON/CLI
- Blacklist users
- Print out whispers for valid routes
- Print out rate limits

## Usage

```sh
# List currencies
poe-arbitrage currencies

# Check for opportunities when trading Chaos Orbs or Exalt Orbs
poe-arbitrage check --currencies chaos,exa

# Check for opportunities in the current hardcore league
poe-arbitrage check --currencies chaos,exa --hard

# Check for opportunities with profit of 5 Chaos
poe-arbitrage check --profit "5 chaos"

# Check for opportunities and fetch latest numbers
poe-arbitrage check --currencies chaos,exa --latest
```

## Open Questions

- How will the stack size affect trades?
  - i.e. Exalt Orb for Chaos Orb where one trade window is not big enough.
  - Current limiting trades to `5 * 12 * stack size`
- What are the API rate limits for POE exchange API?
- Should the same user be traded with multiple times?
- Profitability of flipping inefficiently priced items or gems.
