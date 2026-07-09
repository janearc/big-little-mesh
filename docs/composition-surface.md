# Composition surface

Every unit of composed work on the mesh reduces to two messages in `bento.v1`
(`proto/bento/v1/composition.proto`): a **recipe**, carried by the bento, and an
**intent**, submitted by the caller. This document records the doctrine those
messages encode and the pieces deliberately left out.

## The doctrine: the bento carries the recipe

A caller submits an `Intent` — a recipe reference, an input bento reference, an
identity, a hop bound — and nothing else. The caller never orchestrates. The bento
carries its `Recipe`; each pipeline that holds the bento reads the recipe and works
out the next step itself. No caller stands outside the mesh driving hops, and no
bento is dragged along a wire of fixed stops. This is the "describes what it wants
to become" model from [pipelines.md](pipelines.md#composition-and-destinations),
now on the wire.

## Capability, never a service name

A `Step` names a **capability** — what KIND of work it is (transcription,
text-clean, image-generation) — and MUST NOT name a service. The mesh resolves a
capability to a citizen by discovery at each hop, the same way `model.v1` selects a
model family by capability rather than by a concrete model's name. Naming a service
would prescribe a route; prescribed routes are forbidden. A step's `input` and
`output` are `frood.v1.ContractRef`s — the same RecordNameStrategy subjects a frood
claims in its `ContractDescriptor` — so resolution is checkable: the citizen a
capability resolves to must claim the contracts the step declares.

## Hop bound and attribution

Composition can loop: pipeline A's output submits an intent that eventually reaches
pipeline A again. Two fields on `Intent` handle this, and they do different jobs:

- `hop_bound` **prevents** the loop. Every intent derived from another carries its
  parent's bound minus one. An intent arriving with a bound of zero MUST be
  refused, loudly — never silently dropped. A runaway composition fails visibly.
- `parent_intent_id` **attributes** the loop. Each derived intent names its parent;
  a root intent's is empty. When a bound trips, walking the chain names the recipe
  and submitter that spiraled, not merely the fact that something did.

Prevention without attribution leaves nothing to fix. The bound stops the damage;
the chain tells you whose recipe to repair.

`intent_id` is the idempotency key: the same intent delivered twice is one fact,
not two — the same rule `BentoLifecycleEvent.event_id` already carries.

## Convergence, minimally

`Convergence` declares how the mesh knows a recipe's bento is done or failed: an
`acceptance` check (the same runnable shape a banchan asset carries; unset means
the bento's default completeness rule) and `max_passes`, the bound on the one
sanctioned cycle — a pipeline's own convergence loop. What happens *between*
passes — retry biasing, model swaps, escalation — is the convergence policy, and
it is undefined. It is deferred, not smuggled in half-formed.

## Deferred, by name

- **Intent lifecycle state enum.** An intent will have states; they land with the
  routing implementation, once intent transport is ratified.
- **Intent transport.** Bus versus dataprovider is an open question. Nothing in
  `composition.proto` presumes either.
- **Convergence policy.** The between-passes behavior, per above.
- **Recipe residency.** `Intent.recipe_id` is a reference; where recipes live and
  how the reference resolves is not defined here.
- **Speedy binding.** The mesh's near-realtime capability is called *speedy* — it
  is close to realtime, not realtime, and it is deliberately not called streaming.
  A speedy session will later bind to a bento and its segments to banchans; that
  binding is a follow-up, and no `speedy.v1` package exists yet.

## Why this document exists

Recorded 2026-07-09 as ADR-0002: the composition surface was the largest deferral
in `bento.proto`'s header, and field numbers are permanent — the doctrine behind
them (recipe on the bento, capability over service name, bounded and attributable
derivation) needs to be readable without commit archaeology.
