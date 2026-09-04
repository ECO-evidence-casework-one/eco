# ECO — Deferred Audio Accessibility Requirement

**Date:** 6 August 2026  
**Status:** future Product Lab requirement; not part of the current recovery candidate

## Product requirement

ECO should eventually support:

1. proper keyboard, Narrator and NVDA access;
2. optional offline spoken guidance;
3. optional calm interface sounds;
4. later push-to-talk local voice input.

## Controlling principles

- Screen-reader compatibility comes before ECO's own spoken guide.
- Voice guidance is optional and must not replace accessible names, roles, states, focus order or keyboard operation.
- No important information may be communicated by sound alone.
- Spoken guidance must be interruptible, pausable and repeatable.
- Draft, checked, stopped, rejected and unavailable AI states must be spoken distinctly.
- Guidance about the interface must come from deterministic application state, not improvised model text.
- The system must remain fully offline and use only approved free/open-source components and clearly licensed voice assets.
- ECO speech must not talk over a user's screen reader without explicit user choice.

## First bounded future slice

When the main recovery candidate is stable, the first audio slice should only:

- announce the active Matter and current area;
- read focused-control help on request;
- announce evidence-processing status;
- read selected checked text or a checked AI answer;
- pause, repeat and stop speaking immediately.

## Deferred items

These are not authorised for the current recovery candidate:

- full spoken step-by-step navigation;
- always-on listening;
- voice-controlled consequential actions;
- dictation saved without review;
- speech-engine or voice-model bundling;
- new audio settings screens;
- final voice selection or licensing decision.

## Gate effect

None. This record preserves the requirement without changing RC2 development scope, application source or release status.
