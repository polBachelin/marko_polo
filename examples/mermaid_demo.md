# Mermaid Demo

This file is a visual smoke test for the Mermaid rendering feature in the `marko` markdown reader. It mixes several Mermaid diagram types (ER, flowchart, sequence) with regular prose, headings, lists, and a plain fenced code block — so we can confirm that Mermaid blocks render as diagrams while ordinary code blocks continue to render with normal syntax highlighting.

## 1. Entity-Relationship Diagram

The ER diagram below models the add-on / provider / booking domain. It exercises Mermaid's `erDiagram` syntax including cardinalities, labeled relationships, primary/foreign keys, nullable columns, and inline comments on fields.

```mermaid
erDiagram
    PROVIDER ||--o{ ADDON_OFFERING : fulfills
    ADDON ||--o{ ADDON_OFFERING : "categorized as"
    ADDON ||--o{ ADDON_PRICING : "priced by"
    ADDON ||--o{ ADDON_INVENTORY : "capped by"
    SPACE ||--o{ ADDON_OFFERING : offers
    LISTING ||--o{ ADDON_OFFERING : offers
    BOOKING ||--o{ BOOKING_ADDON : purchases
    ADDON ||--o{ BOOKING_ADDON : "frozen ref"
    PROVIDER ||--o{ BOOKING_ADDON : "frozen ref"
    BOOKING_ADDON ||--|| ADDON_FULFILLMENT : tracks
    BOOKING_ADDON ||--o{ ADDON_PAYOUT : settles
    ADDON ||--o{ ADDON_REFUND_RULE : "governs cancel"

    PROVIDER {
        uuid id PK
        string name
        string type "host | partner | drivestay"
        string stripe_connected_account_id "nullable"
        string contact_phone
        string contact_email
    }
    ADDON {
        uuid id PK
        string code "ev_cart_ride | ev_charging | car_wash"
        string name
        string description
        string fulfillment_mode "phone | api | dashboard | none"
        string pricing_model "fixed | event_date | dynamic"
        bool requires_inventory
        string tax_treatment "hook only"
    }
    ADDON_OFFERING {
        uuid id PK
        uuid addon_id FK
        uuid provider_id FK
        uuid listing_id "nullable"
        uuid space_id "nullable"
        bool bundled "true=included, false=opt-in"
        bool active
    }
    ADDON_PRICING {
        uuid id PK
        uuid addon_id FK
        date event_date "nullable"
        string venue "nullable, partition hook"
        bigint amount_cents
        timestamp effective_from
        timestamp effective_until
    }
    BOOKING_ADDON {
        uuid id PK
        uuid booking_id FK
        uuid addon_id FK
        uuid provider_id FK "frozen"
        bigint amount_cents "frozen"
        bool bundled "frozen"
    }
    ADDON_FULFILLMENT {
        uuid id PK
        uuid booking_addon_id FK
        string status "pending|notified|confirmed|fulfilled|failed"
        timestamp notified_at
        int notification_attempts
        jsonb metadata
    }
    ADDON_PAYOUT {
        uuid id PK
        uuid booking_addon_id FK
        uuid provider_id FK
        bigint amount_cents
        string method "stripe_connect | host_wallet | manual"
        string stripe_transfer_id "nullable"
        string status "pending | sent | reconciled"
    }
    ADDON_INVENTORY {
        uuid id PK
        uuid addon_id FK
        date event_date
        string venue
        int capacity
        int consumed
    }
    ADDON_REFUND_RULE {
        uuid id PK
        uuid addon_id FK
        string rule_type "non_refundable | full | proportional"
    }
```

A few things to look for in the rendered ER diagram:

- All ten entities are present and laid out without overlap.
- Relationship labels (e.g. "frozen ref", "governs cancel") are readable.
- The `||--||` (one-to-one) edge between `BOOKING_ADDON` and `ADDON_FULFILLMENT` is visually distinct from the `||--o{` (one-to-many) edges.

## 2. Booking Lifecycle Flowchart

A simple top-down flowchart that walks a booking through its states. It tests `flowchart TD`, decision nodes, and a cancelled branch.

```mermaid
flowchart TD
    A[Guest submits booking] --> B{Auto-approve?}
    B -- Yes --> C[Pending payment]
    B -- No --> D[Awaiting host review]
    D -->|Host approves| C
    D -->|Host declines| X[Cancelled]
    C -->|Payment succeeds| E[Confirmed]
    C -->|Payment fails / times out| X
    E --> F[Completed]
    E -->|Guest or host cancels| X
```

## 3. Payment Sequence Diagram

This sequence diagram shows the happy-path payment handshake between the guest, our API, Stripe, and the host. It exercises `sequenceDiagram`, participants, and a mix of sync and async messages.

```mermaid
sequenceDiagram
    participant G as Guest
    participant A as API
    participant S as Stripe
    participant H as Host
    G->>A: POST /bookings (create)
    A->>S: Create PaymentIntent
    S-->>A: client_secret
    A-->>G: 200 OK + client_secret
    G->>S: Confirm payment (client SDK)
    S-->>A: webhook payment_intent.succeeded
    A->>H: Push notification "New booking confirmed"
```

## 4. Plain Code Block (Not Mermaid)

The fenced block below is a regular Go snippet. It must render as syntax-highlighted source code, **not** as a Mermaid diagram. If it gets handed to the Mermaid renderer it will either fail to parse or render as garbage — that's the bug we're guarding against.

```go
package main

import "fmt"

func main() {
    statuses := []string{"pending", "confirmed", "completed"}
    for i, s := range statuses {
        fmt.Printf("%d: %s\n", i, s)
    }
}
```

And for good measure, a second non-mermaid block in a different language:

```javascript
const states = ["pending", "confirmed", "completed", "cancelled"];
const isTerminal = (s) => s === "completed" || s === "cancelled";

for (const s of states) {
  console.log(`${s} -> terminal? ${isTerminal(s)}`);
}
```

## Checklist

When viewing this file in `marko`, verify:

- [ ] The ER diagram renders as a diagram (not as raw text).
- [ ] The flowchart renders with arrows and decision diamonds.
- [ ] The sequence diagram shows the four participants and arrows between them.
- [ ] The Go and JavaScript blocks render as syntax-highlighted code, not as diagrams.
- [ ] Headings, paragraphs, and the bullet list around the diagrams render normally.
