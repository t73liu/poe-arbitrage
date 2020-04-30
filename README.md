## Path of Exile Arbitrage

CLI to detect arbitrage opportunities by taking advantage of inefficient
bid-ask spreads in the [Bulk Item Exchange](https://www.pathofexile.com/trade/exchange/Delirium).
This tool takes advantage of the unofficial POE exchange API.

## Feature

- Detect opportunities with a minimum of `N` online users.
  - Decreases potential transaction costs and opportunity costs of holding
    non-liquid currencies.
- Detect opportunities with a minimum of `N` profit (dependent on trading pair)
- Set initial capital via JSON/CLI
- Limit trade pairs via JSON/CLI
- Limit per trade volume via JSON/CLI

## Usage

```sh
# List currencies
arbitrage currencies

## Check for opportunities when buying/selling any currency
arbitrage check

# Check for opportunities with profit of 5 Chaos
arbitrage check -profit 5 -unit chaos

# Check for opportunities when selling Chaos Orbs or Exalt Orbs
arbitrage check -sell chaos,exa

# Check for opportunities when selling Chaos and buying Exalt/Alchemy/Alteration
arbitrage check -sell chaos -buy exa,alch,alt

# Check for opportunities with trade volume limit of 100 Chaos or 2 Exalt
arbitrage check -sell chaos,exa -limit 100,2
```

## Open Questions

- How will the stack size affect trades?
  - i.e. Exalt Orb for Chromatic Orb where one trade window is not big enough.
- What are the API rate limits for POE exchange API?
