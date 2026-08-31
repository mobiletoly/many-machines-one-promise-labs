# Practice 05: What May the Product Claim?

One purchase intent accumulates evidence from independent payment, inventory,
and return authorities. You will review that evidence at successive cuts,
apply the declared product contract, and decide which proposed claims are
supported at each cut.

The product policy is input. Do not invent a more convenient meaning for a
mixed result, and do not let a workflow record replace the authority that owns
an external effect.

## Declared property

Every supported product claim must follow from the evidence available at its
named cut and from a declared contract rule. Later evidence may change the
current set of supported claims, but it does not rewrite an earlier cut.

A rejected corrective attempt does not prove that the correction obligation
is unsatisfied. That stronger conclusion requires complete terminal evidence
for every satisfaction path admitted by the contract.

## Prerequisites and platforms

Install Go 1.23 or newer. The complete verifier requires a POSIX shell and
supports macOS and Linux. Run commands from the repository root.

No HTTP server, database, broker, container runtime, or transaction manager is
required.

## Inspect the dossier

Read:

```text
practices/05-what-may-the-product-claim/contract.json
practices/05-what-may-the-product-claim/workflow-design.json
practices/05-what-may-the-product-claim/authority-evidence.jsonl
practices/05-what-may-the-product-claim/claim-candidates.json
practices/05-what-may-the-product-claim/starter/product-claim-review.json
```

`contract.json` declares the business rules and the propositions asserted by
the candidate claims. Your job is to apply those rules, not write new product
policy. `workflow-design.json` also contains four designs to test against the
stronger all-or-nothing terminal contract.

Do not inspect `solution/product-claim-review.json` yet.

## Commands

Run the starter review. It must fail without printing the reference claim
envelope:

```sh
GOWORK=off go run ./practices/05-what-may-the-product-claim/cmd/replay \
  -review practices/05-what-may-the-product-claim/starter/product-claim-review.json
```

Copy the starter and fill in your review:

```sh
cp practices/05-what-may-the-product-claim/starter/product-claim-review.json \
  /tmp/practice-05-product-claim-review.json
${EDITOR:-vi} /tmp/practice-05-product-claim-review.json
```

For every evidence cut, record:

- the state of each declared operation and the evidence that supports it;
- the contract rules established at that cut;
- whether the return obligation is triggered, unresolved, satisfied, or
  proved unsatisfied;
- each proposed claim's verdict, reason, and evidence.

Then evaluate every architecture under the stronger contract. The dossier asks
whether the architecture fits; it does not ask you to select two-phase commit.

Verify your artifact:

```sh
./scripts/verify-practice-05.sh /tmp/practice-05-product-claim-review.json
```

After it passes, compare it with the bundled reference:

```sh
GOWORK=off go run ./practices/05-what-may-the-product-claim/cmd/replay \
  -review practices/05-what-may-the-product-claim/solution/product-claim-review.json
```

Run the scope boundaries:

```sh
GOWORK=off go test -count=1 \
  ./practices/05-what-may-the-product-claim/replay \
  -run '^TestBoundary'
```

Run the repository check:

```sh
./scripts/check-practice-05.sh
```

## Boundary

A passing review proves only that these claims follow from this declared
contract and evidence. It does not give the application authority to revise an
external outcome or invent business policy.

The stronger-contract review can prove that an architecture fails the required
terminal boundary. It does not automatically require distributed atomic
commit: one local atomic authority may be smaller, a required participant may
be unable to prepare, and a design that always aborts violates the declared
healthy-case progress requirement.
