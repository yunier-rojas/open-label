# Open Label

**An experiment in applying cryptography to shipping labels.**

This project explores how entities that ship parcels — such as businesses, fulfillment providers, or third-party logistics (3PLs) — could generate their own shipping labels that are **cryptographically verifiable** and **interoperable across carriers**, without relying on centralized APIs or proprietary formats.

---

## The Experiment

What if labels worked like signed digital documents instead?

In this model, the shipping entity generates a label once, signs it, and shares it. Carriers and third parties can then consume and verify the label’s authenticity and integrity, rather than being the sole issuers.

---

## Key Concepts

1. **Distributed, Cryptographically Signed Labels**

   * Shipping entities generate their own public–private key pairs.
   * Public keys are published so carriers and partners can verify authenticity.

2. **Readable + Verifiable Data Format**

   * Labels use **YAML** for both human and machine readability.
   * Public claims are signed to prevent tampering.

3. **Namespaces for Clarity & Extensibility**

   * `i_`: Issuer (entity signing the label)
   * `s_`: Sender / shipping entity details
   * `r_`: Recipient / destination details
   * `p_`: Package information (weight, dimensions, etc.)
   * `v_`: Service requirements (e.g., insurance, signature)
   * `u_`: URLs for callbacks or restricted resources
   * `x_`: Metadata for versioning, timestamps, and the final signature (`x_sig`)

---

## How could it work?

1. **Label Generation**

   * The shipping entity prepares a YAML payload with shipment details.
   * The payload is cryptographically signed with their private key.
   * The signed YAML is embedded in a QR code.
   * A shipping label is created with both the QR code and human-readable info.

2. **Parcel Shipment**

   * The label is attached to the parcel and handed to the carrier.

3. **Carrier Verification**

   * The carrier scans the QR code to retrieve the payload.
   * Using the shipping entity’s public key, they verify:

      * That the label hasn’t been tampered with.
      * That the issuer is authentic.
   * Sensitive claims can be encrypted and decrypted with shared keys if needed.

---

## Food for Thought

This is not a finished product — it’s a **conversation starter** about decentralization and logistics.

* How could issuers and carriers securely exchange **private claims** or status updates?
* Could this complement existing logistics standards?
* What about payments, contracts, pickups, etc?

