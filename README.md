# dreamtree

the **data-sovereign network**: a user-owned, append-only attestation store and the
protocol around it. The wallet on it is called **roots**. Both names are lowercase, always.

> **Status: the wallet is live.** [roots](https://github.com/blong-dev/roots) runs in
> production at [`id.dreamtree.org`](https://id.dreamtree.org), holding real wallets with
> real credentials, each anchored to a publicly resolvable DID. The protocol and data model
> in this repo are the design it grew from, still being derived one layer at a time. Specs
> here are working documents, not frozen contracts.

## The atom

Everything rests on one immutable unit — the **observation**:

```
Observation = (C, A, T, S, σ)
```

a referent/event `C`, an attesting entity `A`, an observation time `T`, a
statement `S`, and a proof `σ`. In data, a thing does not exist until it is
observed. The log of observations is the **sole ground truth**; identity,
things, edges, gravity, and value are all *derived, reprocessable projections*
over that log. Legitimacy and referent-time are derived, never stored.

## The chain — running

This repo now holds the chain itself, not just its design. A Cosmos SDK + CometBFT
app-chain (per the resolved decisions in [`protocol-spec.md`](protocol-spec.md)):
photons as the native unit, `dream1…` addresses, instant finality.

> **Honest status: v0 devnet.** One validator, run by us — internally we call this
> "a really over-engineered database," and we will not claim decentralization we
> have not earned. The validator-set schedule (solo → federated → permissioned-open
> → open) is in the spec, published rather than promised. No ICO, no token sale,
> ever; photons are work-accounting, not speculation.

```
# build & run a local devnet (Go 1.23+)
make install
bash scripts/init.sh
dreamtreed start --minimum-gas-prices 0photon
```

The value layer (`x/seeds` commitments anchoring the wallet's records, attestations,
reputation, licenses) lands module by module, in the open.

## The pieces

- **[roots](https://github.com/blong-dev/roots)** — the wallet, live. Consent-gated reads,
  provenance-scoped grants, verify-at-read tiers, append-only correction, signed full export.
- **This repo** — the protocol and data model the wallet implements, and the layers still
  ahead (the attestation economy, the chain, the network).

## Design docs

- [`data-model.md`](data-model.md) — the atom and the layered data model
- [`data-types.md`](data-types.md) — concrete types over the model
- [`protocol-spec.md`](protocol-spec.md) — the dreamtree protocol
- [`wallet-spec.md`](wallet-spec.md) · [`wallet-v0.md`](wallet-v0.md) — the wallet design roots was built from
- [`credential-profile.md`](credential-profile.md) — the W3C VC profile issuers follow
- [`parameters.md`](parameters.md) — system parameters
- [`VISION.md`](VISION.md) — where this is going
- [`manifesto/`](manifesto/) — the why

## License

[AGPL-3.0](LICENSE). The wallet must be uncompromised — the data-sovereignty
fight requires it.
