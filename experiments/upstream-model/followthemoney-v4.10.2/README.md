# Selected FollowTheMoney schemas for ECO integration

Upstream: `opensanctions/followthemoney`

Release: `v4.10.2`

Exact commit: `b9418ecd32bd60dd09c261134464860a0082ffb7`

Licence: MIT (see `LICENSE` in this directory).

These files are vendored **unchanged** from the exact upstream release for schema/model evaluation only:

- `Person.yaml`
- `Organization.yaml`
- `Event.yaml`
- `Document.yaml`
- `Email.yaml`

ECO does not adopt FollowTheMoney as its trust model. These schemas provide a mature investigative vocabulary/interoperability target. ECO's own `InformationOrigin`, `SourceLink`, preserved-original identity, user-confirmation, conflict, uncertainty and recovery fields remain controlling.

The intended adapter maps selected fields into ECO entities while preserving exact source links and review state. No automatic entity assertion is authorised merely because a parser or model proposes a value.
