# 🛰️ AWS WebSocket Chat Demo

This project demonstrates a basic real-time messaging system using **AWS API Gateway WebSocket**, **AWS Lambda**, **DynamoDB**, and a **Golang WebSocket client**. It allows multiple clients to connect and receive broadcast or direct messages in real-time.

---

## 🧱 Project Structure

### 1. `ws-starter.yaml`
A CloudFormation template that creates:

- A **DynamoDB** table for tracking WebSocket connections
- Four **Lambda functions**:
  - `$connect`: Adds a new connection ID to DynamoDB and optionally sends it to the client
  - `$disconnect`: Removes the connection ID when the client disconnects
  - `sendServer`: Send message to the peer server
- **IAM roles** and policies for the Lambda functions

### 2. `main.go`
A **Golang WebSocket client** that:
- Connects to the WebSocket API Gateway
- Sends an initial `getConnectionId` request
- Maintains the connection with regular pings
- Listens for incoming messages and prints them in the format:

```
📨 Message from [username]: [message]
```

---

## 🚀 Getting Started

### 🛠 Deploying the Stack

1. Make sure you have AWS CLI configured and run:

```bash
aws cloudformation deploy \
  --template-file ws-starter.yaml \
  --stack-name websocket-chat-demo \
  --capabilities CAPABILITY_IAM
```

2. After deployment, create an **Amazon API Gateway WebSocket API** and set the routes:

| Route             | Integration Target                |
|-------------------|-----------------------------------|
| `$connect`        | ConnectHandler Lambda ARN         |
| `$disconnect`     | DisconnectHandler Lambda ARN      |
| `sendServer`      | SendServerHandler Lambda ARN      |
| `getConnectionId` | GetConnectionIdHandler Lambda ARN |

3. Deploy the WebSocket API to a stage (e.g., `production`) and grab the **WebSocket endpoint URL**.

---

## 🔄 Message Flow

1. 🔎 **Get your connection ID**  
   After your client connects to the WebSocket server, send this message:

   ```json
   { "action": "getConnectionId" }
   ```

   The server will reply with your connection ID. You can share it with another client for direct messaging.

---

## 💻 Example Usage with wscat

### 🧪 Step-by-step:

#### 1️⃣ Client 1 connects as a `cli` server:
```bash
wscat -c 'wss://6we5i7fnq5.execute-api.ap-southeast-1.amazonaws.com/production?type=cli&cliId=6f3a8b0b-1f7f-4681-ae94-2b213f95f8a0' -H "Authorization: Allow"
```

#### 2️⃣ Client 2 connects as an `mpc` server:
```bash
wscat -c 'wss://6we5i7fnq5.execute-api.ap-southeast-1.amazonaws.com/production?type=mpc&mpcId=b61c9f9e-13b8-4c87-a750-16e1cfdabe9c' -H "Authorization: Allow"
```

#### 3️⃣ Client 2 sends a message to Client 1:
```json
{
  "action": "sendServer",
  "sourceId": "b61c9f9e-13b8-4c87-a750-16e1cfdabe9c",
  "cliToMpc": { "6f3a8b0b-1f7f-4681-ae94-2b213f95f8a0": "b61c9f9e-13b8-4c87-a750-16e1cfdabe9c" },
  "operationType": "signing",
  "message": {
    "txId": "abc123",
    "signature": "0xdeadbeef"
  }
}
```

🔁 **Client 1 receives:**
```json
{
  "operationType": "signing",
  "from": "IgniterC56D",
  "message": {
    "txId": "abc123",
    "signature": "0xdeadbeef"
  }
}
```

#### 4️⃣ Client 1 replies back to Client 2:
```json
{
  "action": "sendServer",
  "sourceId": "6f3a8b0b-1f7f-4681-ae94-2b213f95f8a0",
  "operationType": "fullSig",
  "message": {
    "txId": "abc123",
    "signature": "0xdeadbeef"
  }
}
```

🔁 **Client 2 receives:**
```json
{
  "operationType": "fullSig",
  "from": "cli123",
  "to": "IgniterC56D",
  "message": {
    "txId": "abc123",
    "signature": "0xdeadbeef"
  }
}
```

---

## 💬 Running the Golang Client

### 🔧 Install dependencies

This uses the [coder/websocket](https://github.com/coder/websocket) Go library. Make sure you have Go installed, then:

```bash
go mod tidy
```

### ▶️ Run the client

Update the `wsURL` in `main.go` with your WebSocket endpoint:

```go
const wsURL = "wss://<your-api-id>.execute-api.<region>.amazonaws.com/production/"
```

Then run:

```bash
go run main.go
```

---

## 📡 Sending Messages

You can trigger the `sendMessage` Lambda manually via:

```bash
aws lambda invoke \
  --function-name <SendMessageHandler Lambda Name> \
  --payload '{"action": "sendmessage", "message": "Hello from CLI!", "apigw_endpoint": "https://<your-api-id>.execute-api.<region>.amazonaws.com/production"}' \
  output.json
```

Or hook it to another Lambda/event to send automatically.

---

## 🧼 Clean Up

To delete all resources:

```bash
aws cloudformation delete-stack --stack-name websocket-chat-demo
```

---

## 📝 License

This project is provided for educational/demo purposes and is not production-hardened. Use at your own risk.

---

## 📋 Real Interaction Example

### 🔌 Client1 connects as `cli` server by launch wsClient.go
```bash
go run wsClient.go
```

### 🔌 Client2 connects:
```bash
wscat -c 'wss://eqm3whvj69.execute-api.ap-southeast-1.amazonaws.com/production?type=mpc&mpcId=IgniterC56D' -H "Authorization: Allow"
```

### 📨 Client2 sends:
```json
{
  "action": "sendServer",
  "sourceId": "IgniterC56D",
  "cliToMpc": { "cli123": "IgniterC56D" },
  "operationType": "PartialSig",
  "message": {
    "accountHash": "accountHash",
    "teamId": "teamId",
    "transactionId": "transactionId",
    "partialSig": "partialSig"
  }
}
```

### 📥 Client1 receives:
```json
{
  "operationType": "PartialSig",
  "from": "IgniterC56D",
  "message": {
    "accountHash": "accountHash",
    "teamId": "teamId",
    "transactionId": "transactionId",
    "partialSig": "partialSig"
  }
}
```

### 📤 Client1 auto sends:
```json
{
  "action": "sendServer",
  "sourceId": "cli123",
  "operationType": "FullSig",
  "message": {
    "accountHash": "accountHash",
    "signatureR": "r",
    "signatureS": "s",
    "signatureV": "v",
    "teamId": "teamId",
    "transactionId": "transactionId"
  }
}
```

### 📬 Client2 receives:
```json
{
  "operationType": "FullSig",
  "from": "cli123",
  "to": "IgniterC56D",
  "message": {
    "accountHash": "accountHash",
    "signatureR": "r",
    "signatureS": "s",
    "signatureV": "v",
    "teamId": "teamId",
    "transactionId": "transactionId"
  }
}
```
