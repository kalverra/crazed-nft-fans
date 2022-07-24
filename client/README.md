# Transactions

The transaction tracker keeps track of all transactions made. It can help with nonce management a bit, but the intent is that the onus of keeping track of which nonce to use is on the caller. The tracker will provide info to help with those decisions, but not automatically determine nonces.
