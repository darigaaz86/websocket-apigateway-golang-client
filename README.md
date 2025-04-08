# 🛰️ AWS WebSocket Chat Demo

This project demonstrates a basic real-time messaging system using **AWS API Gateway WebSocket**, **AWS Lambda**, **DynamoDB**, and a **Golang WebSocket client**. It allows multiple clients to connect and receive broadcast messages in real-time.

---

## 🧱 Project Structure

### 1. `ws-starter.yaml`
A CloudFormation template that creates:

- A **DynamoDB** table for tracking WebSocket connections
- Three **Lambda functions**:
  - `$connect`: Adds a new connection ID to DynamoDB
  - `$disconnect`: Removes the connection ID when the client disconnects
  - `sendMessage`: Scans all active connections and sends a message to each one
- **IAM roles** and policies for the Lambda functions

### 2. `main.go`
A **Golang WebSocket client** that:
- Connects to the WebSocket API Gateway
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
  --template-file cloudformation-template.yaml \
  --stack-name websocket-chat-demo \
  --capabilities CAPABILITY_IAM
```

2. After deployment, create an **Amazon API Gateway WebSocket API** and set the routes:

| Route         | Integration Target             |
|---------------|-------------------------------|
| `$connect`    | ConnectHandler Lambda ARN      |
| `$disconnect` | DisconnectHandler Lambda ARN   |
| `sendMessage` | SendMessageHandler Lambda ARN  |

3. Deploy the WebSocket API to a stage (e.g., `production`) and grab the **WebSocket endpoint URL**.

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
  --payload '{"message": "Hello from CLI!", "apigw_endpoint": "https://<your-api-id>.execute-api.<region>.amazonaws.com/production"}' \
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
