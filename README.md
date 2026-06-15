# DreamTree (DTW)

The **data-sovereign wallet** — a user-owned, append-only attestation store, and
the backbone of the Quorum surfaces (Telekora, the Great Library, Cosmo).

> **Status: early design.** The protocol and data model are being derived from
> first principles, one layer at a time. Specs here are working documents, not
> frozen contracts.

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

## Design docs

- [`data-model.md`](data-model.md) — the atom and the layered data model
- [`data-types.md`](data-types.md) — concrete types over the model
- [`protocol-spec.md`](protocol-spec.md) — the DTW protocol
- [`wallet-spec.md`](wallet-spec.md) · [`wallet-v0.md`](wallet-v0.md) — the wallet itself
- [`parameters.md`](parameters.md) — system parameters
- [`VISION.md`](VISION.md) — where this is going
- [`manifesto/`](manifesto/) — the why

## License

[AGPL-3.0](LICENSE). The wallet must be uncompromised — the data-sovereignty
fight requires it.
