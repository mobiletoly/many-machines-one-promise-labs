# Practice 06: Can This Participant Finish?

This Practice asks the reader to certify one participant's recovery
disposition at five named evidence cuts. It does not implement a transaction
manager or select a distributed commit protocol.

The evidence corpus separates three questions:

1. Did the target participant retain the capability promised by `PREPARED`?
2. What, if anything, does decision authority `D` establish at the cut?
3. Does the retained decision satisfy the fixed participant contract?

The allowed dispositions are:

```text
APPLY_COMMIT
APPLY_ABORT
REMAIN_PREPARED
REPORT_PROTOCOL_BREACH
```

Read these files before running the starter:

```text
contract.json
cases.json
participant-evidence.jsonl
decision-evidence.jsonl
starter/adjudication.json
```

Run the deliberately wrong starter from the repository root:

```sh
GOWORK=off go run ./practices/06-atomic-commit-adjudication/cmd/replay \
  -review practices/06-atomic-commit-adjudication/starter/adjudication.json
```

Copy the starter outside the repository, edit it, and verify it:

```sh
cp practices/06-atomic-commit-adjudication/starter/adjudication.json \
  /tmp/practice-06-adjudication.json
./scripts/verify-practice-06.sh /tmp/practice-06-adjudication.json
```

The verifier derives every disposition from the protocol contract and raw
evidence. It does not select answers by case ID or compare the reader file with
the bundled solution.

Repository maintainers can run the full Practice check:

```sh
./scripts/check-practice-06.sh
```

The human author froze this Practice after independent correction verification.
The complete-public-cut candidate includes it. No release tag, push, or
publication follows from that freeze.
