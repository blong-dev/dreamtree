# DreamTree Vision
*Draft 1 — March 2026 (refined 2026-05-06)*

*This document is the wallet/mission frame, and it now lives in the DTW repo (`blong-dev/dreamtree`) alongside the data model and protocol it serves. The old workbook app (Next.js 15 + Cloudflare Workers, dreamtree.org) is archived at `blong-dev/DreamTree-oldsource` as the dogfood/cannibalization source — not active dev.*

---

## What DreamTree Is

DreamTree is a wallet for who you are. We refer to it as DTW.

Not a financial wallet. An identity wallet — a place where everything verifiable about you lives, in a form you own, that can speak for you without you having to repeat yourself. Drop in your resume. Have a conversation. Grant access to your work documents. The wallet fills itself. You confirm what it gets right. From then on, it represents you.

The wallet holds your skills, your values, your work history, your credentials, your references — already gathered, already verified, ready to travel anywhere you go.

**DTW is the brand and the mission.** The career-coaching workbook that the dreamtree-app currently delivers is no longer the product — it is *a* course that runs on top of the wallet. dreamtree.org's job is the wallet itself, with a course catalog (the workbook + data literacy + "how to use your DTW" + future wallet-aligned courses) running alongside it as the dogfood and the on-ramp.

---

## The Wider Frame — Your Handshake with the Tech World

The education-to-employment handshake is the first instance, not the whole. DreamTree is meant to be the **individual's handshake with the entire tech and AI world** — a *secular* layer (neutral, owned by no platform) that sits between a person and everything that wants their data or attention, and shakes hands on their behalf.

Three things it does:

- **Protects and organizes.** Your verified self, kept by you, in one place — not scattered across a hundred services that each own a sliver and sell it.
- **Harnesses agents to serve you.** Your wallet is the trustworthy, machine-readable context any agent needs to understand you *as you need to be understood* — and then go produce value for you, on whatever surface and in whatever avenue fits. The same verified record powers a job application, a research agent, a negotiation, a course. One harness, every surface.
- **Brings you the best models, cheaply.** Part of the portal is matching individuals to the highest-end models their need calls for, at the lowest cost — democratized frontier capability, brokered *for the person* rather than rented from a platform.

Underneath, **DTW reshapes how information is stored and verified** — provenance-complete, attributed, AI-native — so the thing the early internet promised (find what's true, trust what you find) finally becomes feasible for both people and machines. The current stack — the workbook, Telekora, gnosis — is **scaffolding** toward that.

### Sovereignty is proactive, and it's leverage

Sovereignty isn't waiting for platforms to hand your data back. It's **keeping your own record, first.** DreamTree tracks your side proactively: *I was here, I did this, this was shown to me, at these times* — first-person observations you author, with the capture as proof. Two tiers follow:

- **What you track yourself** is unassailably yours — authored by you, about you, receipted.
- **What platforms derive about you** you can't author — but your own record is the **leverage** to get it. "I can prove I was there and this happened" means that when GDPR (and what comes after it) presses a platform, you hold independent receipts: their disclosure can be audited for completeness, and you negotiate from proof instead of asking nicely.

That's the bet underneath the data-marketplace and attribution use-cases below: **arm individuals with verified receipts ahead of where the law is already walking.** Good faith becomes both sides holding evidence that checks out.

---

## The Problem It Solves

Every time you apply for a job, enroll in a course, or prove you're qualified for something, you start from zero.

Fill out the form. Upload the resume. Wait for the background check. Track down your former boss for a reference. Write the cover letter that explains the same story you've explained a hundred times.

The recruiter reads it for six seconds. Guesses. You either get a call or you don't.

The problem isn't that people are unverifiable. The problem is that verification doesn't travel. Your boss already knows you did the work. Your professor already knows you learned it. Your colleague already witnessed what you built. That knowledge exists — it just lives in someone's head, or in a database you can't access, or behind a phone call that takes three weeks to schedule.

The education-to-employment handshake is the most expensive, most lossy transaction in the economy. Massive investment on both sides. The matching mechanism is: here's a PDF, hope it works out.

The wallet fixes this by making human verification portable.

---

## The Wallet Primitive

The wallet is a personal data store. You own the keys. You control what's in it. You decide what to share and with whom.

What goes in:

- **Skills** — not just labels, but skills in context. Where did you learn this? How well can you do it? What's the evidence?
- **Values** — what energizes you, what drains you, what environments bring out your best work
- **Stories** — structured narrative of moments that define how you think and solve problems
- **Experiences** — your work and education history, with the skills mapped to each
- **Credentials** — diplomas, certifications, completions — issued by institutions and stored as verifiable proofs
- **Attestations** — your boss's signature on a project. A colleague's witness to your contribution. References that travel with you, not names to cold-call
- **Profile** — the headline, the summary, the identity story — synthesized from everything else

How it fills itself:

- **Resume drop** — upload a document, the wallet extracts and structures it
- **Conversation** — talk to an AI. It asks questions, you answer, the profile builds naturally
- **Document access** — grant access to work documents, performance reviews, communications with consent
- **Passive signals** — calendar patterns, work rhythms, collaboration signals — opt-in, transparent, yours
- **Credential receipt** — when an institution issues something, it lands in your wallet automatically
- **Attestation capture** — when your boss signs off on a project in their workflow, that signature becomes a receipt in yours

The workbook — the structured self-reflection exercises — is one intake method. A good one, for people who want the full coaching experience. But not the only one. Not even the primary one. The wallet should fill itself from signals that already exist.

---

## The Use Cases

**Job applications.** You apply once. The wallet autofills every form, every ATS, every HR onboarding system. You stop entering the same data into fifty different boxes. Employers get richer information than a PDF resume. The matching improves. The guesswork shrinks.

**Credential verification.** "Did you really graduate from there?" becomes a lookup, not a phone call. The institution issued the credential into your wallet. Anyone you authorize can verify it instantly. No third-party verification services. No waiting.

**References.** Instead of names to cold-call, your references are already captured — attestations signed by people who worked with you, stored in your wallet, verifiable by anyone you authorize. Your former boss doesn't need to answer another phone call. They already said what they needed to say.

**Data marketplace.** You own your data. If a researcher wants employability patterns, you can sell them access — with consent, with attribution, with payment going to you. Not to a platform that extracted it from your behavior. To you. The GDPR-forward model: advertisers get cleaner data, users get paid, extractive middlemen are eliminated.

**Attribution.** Every contribution you've made — to knowledge, to work, to a project — is traceable to you. This matters more as AI generates more output. The wallet is the substrate for proving what a human did versus what a machine did. And eventually, what a machine contributed that was worth compensating.

---

## Humans and Machines Are the Same Problem

This is where DreamTree connects to something larger.

The wallet isn't just for humans. Machines — AI agents — face the same problem. Cosmo does work. Research, analysis, decisions, builds. That work has value. But none of the provenance travels. It disappears into context windows. There's no portable record of what was contributed, no verification of capability, no receipt.

The wallet is the same primitive whether the subject is a human professional or an AI agent. A portable, verifiable, owned record of contribution and capability.

The economy is about to need this for machines for the same reasons it needs it for humans. Attribution infrastructure serves both. The distinction between "human data wallet" and "machine data wallet" is a distinction that won't hold.

---

## How It Connects to Telekora

DTW is built once and used everywhere. The Telekora engine — tool registry, ontology, connections, smart-content engine, course delivery — embeds DTW as the silent backbone for every learner. Every Telekora user is a DTW instance from day zero, whether they know it or not. Credentials, badges, and attestations land in the wallet as the user moves through courses. When the user is later ready to bring in external attestations or carry credentials to an employer, the wallet is already populated, and the upgrade path is "make explicit what was always there."

The same engine powers two distinct product surfaces:

- **Telekora-the-LMS** (telekora.com) — gamified, fixed-layout, B2B LMS conventions. Sells to managers, MLMs, corporate L&D.
- **dreamtree.org** — sovereign-feeling, anti-gamification, conversational, user-owned aesthetic. Serves wallet management plus the wallet-aligned course catalog. This is where the workbook lives.

Same ontology. Same connections. Same smart-content commitment. Different UX grammar. The workbook is dogfooded against the engine but directed to dreamtree.org's UX, because gamification and "user-owned aesthetic" are not the same product as the LMS view a B2B buyer wants.

---

## What the Old App Got Right

The DreamTree workbook — the extended coaching experience built in 2024-2025 — got something important right even if the implementation was wrong.

**The ontology.** The data structure is nearly complete. It captures the full complexity of human soft data in a way that's rare: skills with mastery and evidence, values broken into work and life dimensions, stories structured with the SOARED framework, experiences mapped to skills, career options scored for coherence and needs-alignment, competency assessments, financial needs modeling, network contacts. This is a genuine intellectual achievement. Most attempts at human profiles flatten people into keywords. This one doesn't.

**The workbook content.** Three parts — Roots (self-knowledge), Trunk (connecting past to future), Branches (reaching into the world) — are a synthesis of what career coaches, therapists, mentors, and practitioners across traditions have found meaningful. It's not a guess at what matters. It's a collage of observed expertise.

**The networking and outreach model.** The workbook has explicit content on how professional relationships form, why they matter, and what good outreach looks like. This knowledge is directly applicable to tools like Scout and any system that touches professional communication. It's grounded in something real, not a template.

**The vision.** The PRINCIPLES.md is a serious document. Data sovereignty enforced by cryptography, not promises. A contribution economy, not an attention economy. Open source as survival, not generosity. Trust-capturing, not trustless. Receipts, not transactions. These are the right principles for the right problem.

---

## What the Old App Got Wrong

**Form-first.** The implementation asked people to fill out forms — elaborate, thoughtful, well-designed forms, but forms. No one will spend weeks on self-reflection exercises, no matter how good the underlying framework is. The activation energy was too high. The wallet should fill itself from signals that already exist. The form is a last resort, not the primary channel.

**Complexity as care.** The app grew because of genuine care about capturing human complexity. But the complexity leaked into the interface. A tool that respects human complexity shouldn't feel complex to use. The workbook should feel like a conversation with a thoughtful friend, not like filling out a government application.

**The product was the workbook.** The workbook is the research apparatus — the tool for understanding what data matters, for building and validating the ontology. It's not the product. The product is the wallet. The workbook is one way to fill it.

---

## The Build Target

We will know DreamTree is ready to build when this vision reads like a simple novel. Not a spec. Not a feature list. A story that a non-technical person can read and say: yes, I want that.

The minimum thing that proves the concept:

1. Drop in a resume. Have a conversation. The wallet builds a profile.
2. You review it, confirm what's right, correct what's wrong.
3. The wallet autofills one real job application.

That's it. That's the proof. If a person can go from "I have a resume" to "my application is filled out" without entering data manually, the core value is proven. Everything else — credentials, attestations, marketplace, attribution — builds on that moment.

The workbook remains available for people who want the full experience. The conversation is the default. The form is the fallback.

---

## The Larger Picture

DreamTree is not a career app. It's attribution infrastructure.

The gap between human contribution and economic reward is not natural. It's a function of missing provenance. We don't have a reliable way to say: this person created this value, at this time, with this evidence. So reward flows to whoever can claim the contribution, not whoever made it.

The wallet closes that gap. Not completely. Not immediately. But structurally.

When contribution is receipted at creation — when your boss signs the attestation, when the institution issues the credential, when the colleague witnesses the work — extraction becomes visible. You can trace the delta between who created and who captured. The receipt exists. The value was created. Attribution is verifiable. Compensation becomes a lookup, not a negotiation.

This is TCP/IP for human contribution. Not a product. A protocol. The product proves the protocol works. Then the protocol becomes infrastructure. Then it's no longer owned by anyone — it belongs to everyone who builds on it.

That's the long bet.

---

*Last updated: 2026-06-16 (added "The Wider Frame" — the handshake with the tech world, the agent-harness, proactive two-tier sovereignty, and cheap model access).*
*Status: DTW is being lifted out of the workbook app. The shipped dreamtree-app's components redistribute as Telekora-engine tool registry entries, the v0 wallet ontology (the 40-table schema), and dreamtree.org's sovereign delivery surface. Workbook content survives as the inaugural course in the dreamtree.org catalog, alongside data-literacy and DTW-education courses still to be built.*
