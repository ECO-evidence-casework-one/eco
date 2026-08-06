# ECO Product Lab — Current Position

**Date:** 6 August 2026  
**Status:** documentation-only Product Lab record

## Purpose of this record

This file records the current user-facing product position without changing application source, development scope, workflows, packaging, release gates or model/runtime integration.

## Product Lab role

The Product Lab defines and reviews:

- user experience;
- screen hierarchy and terminology;
- accessibility and cognitive load;
- AI behaviour as experienced by the user;
- product acceptance and regression decisions.

Development builds the application. Independent inspection verifies technical and release claims.

## Current product direction

The controlling recovery direction is:

> Preserve the recovered V38 application, integrate V39 engines incrementally, transplant only successful runtime/packaging work, and apply Product Lab improvements without another wholesale interface replacement.

The rejected V40 replacement shell is not the application parent.

## RC1 position

`ECO-RECOVERY-20260805-RC1` established that:

- the recognisable V38 interface was recovered;
- genuine local Qwen2.5 1.5B inference ran on the Acer baseline;
- the rejected V40 shell was absent;
- conversation continuity and local AI startup were visible;
- the application remained a private synthetic test candidate.

RC1 was not accepted as the next successful ECO version because:

- explicit word-count and exact-text instructions were not obeyed;
- the reply verifier accepted those non-compliant answers;
- the complete Matter → Evidence → OCR → Current Position → source answer → exact source → Stop → close/reopen journey remained incomplete;
- active Matter hierarchy and several user-facing states remained unfinished;
- current accessibility and Acer end-to-end evidence remained incomplete.

## RC2 position

RC2 is under development.

The Product Lab is not adding feature scope during that work.

The next candidate should preserve V38 continuity and correct the bounded RC1 defects while proving one complete synthetic Matter journey.

## Deferred future requirement

Audio accessibility remains a future Product Lab requirement:

- proper Narrator/NVDA and keyboard support first;
- optional offline spoken guidance;
- optional interface sounds;
- later push-to-talk voice input.

This is deferred and is not part of the current recovery candidate.

## Release and evidence boundary

This record does not approve:

- real or sensitive evidence;
- a public executable;
- signing;
- a GitHub Release;
- public, institutional, legal, medical or forensic deployment;
- accessibility, security or compliance claims.
