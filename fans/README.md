# Fans and Their Leader

The fan leader controls all the NFT fans, creating, destroying, activating and deactivating fans.

## Fan Behavior

```mermaid
flowchart LR
    Cre[Create] --> Conf[Configure]
    Conf --> Act[Activate]
    Act --> SG[Send/Guzzle]
    SG --> Tx[Transaction]
    Tx --> W[Wait]
    W --> Timeout[Timeout]
    Timeout --> BG[BumpGas]
    BG --> SG
    W --> Suc[Success]
```
