# AEIB v0.1 Security Boundary

AEIB is a defensive synthetic-input benchmark. It is not an exploit kit.

## Allowed corpus properties

- malformed but inert document/container structures;
- duplicate and nested archive/email objects;
- path strings used to test rejection/quarantine behavior;
- invalid encodings and byte sequences;
- deterministic random bytes;
- modest bounded compression-ratio fixtures;
- source-mutation pairs for controlled local timing tests.

## Excluded from v0.1

- live malware;
- executable persistence payloads;
- credential theft;
- command-and-control behavior;
- destructive archive bombs;
- personal/sensitive evidence;
- third-party exploit chains or undisclosed exploit payloads;
- automatic testing of systems not owned or explicitly authorized.

## Publication rule

A future corpus fixture that crosses from malformed-input testing into exploit reproduction must be separately reviewed before publication. The default is to publish the minimal safe regression form after the affected software is fixed or the responsible maintainer agrees to disclosure.
