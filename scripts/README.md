# Helmjet Atlas Scripts

Utility scripts for setting up and managing Helmjet Atlas.

## Seed MongoDB with Mock Data

Populate your MongoDB instance with realistic sample data for testing and demonstration purposes.

### What Gets Seeded

**Microservices (6):**
- api-gateway
- auth-service
- user-service
- order-service
- payment-service
- notification-service

**NATS Streams (5):**
- events
- orders
- payments
- users
- notifications

**Consumers (6):**
- api-gateway-consumer
- order-processor
- payment-processor
- user-sync
- email-notifier
- sms-notifier

**Links (11):**
- Service-to-stream relationships (5)
- Consumer-to-service relationships (6)

### Prerequisites

- MongoDB running and accessible
- Go 1.21+ installed
- MongoDB URI accessible (default: `mongodb://localhost:27017`)

### Usage

#### PowerShell (Windows)

```powershell
cd scripts
.\seed-mongodb.ps1
```

Or with custom MongoDB URI:

```powershell
.\seed-mongodb.ps1 -MongoUri "mongodb://my-mongo-server:27017" -MongoDb "my-database"
```

#### Bash (macOS/Linux)

```bash
cd scripts
./seed-mongodb.sh
```

Or with custom MongoDB URI:

```bash
export MONGO_URI="mongodb://my-mongo-server:27017"
export MONGO_DB="my-database"
./seed-mongodb.sh
```

#### Direct Go Command

```bash
cd scripts
MONGO_URI=mongodb://localhost:27017 MONGO_DB=helmjet-atlas go run seed-mongodb.go
```

### Environment Variables

- `MONGO_URI` - MongoDB connection string (default: `mongodb://localhost:27017`)
- `MONGO_DB` - Database name (default: `helmjet-atlas`)

### What the Script Does

1. **Connects** to MongoDB using the provided URI
2. **Clears** existing collections to start fresh
3. **Creates** microservices with realistic attributes
4. **Creates** NATS streams with subject patterns
5. **Creates** consumers linked to streams
6. **Creates** relationships between services and streams
7. **Creates** relationships between consumers and services
8. **Prints** summary of created entities

### Output Example

```
🌱 Seeding MongoDB with mock data...
MongoDB URI: mongodb://localhost:27017
Database: helmjet-atlas

Created microservice: api-gateway
Created microservice: auth-service
Created microservice: user-service
Created microservice: order-service
Created microservice: payment-service
Created microservice: notification-service
Created stream: events
Created stream: orders
Created stream: payments
Created stream: users
Created stream: notifications
Created consumer: api-gateway-consumer
Created consumer: order-processor
Created consumer: payment-processor
Created consumer: user-sync
Created consumer: email-notifier
Created consumer: sms-notifier
Created service-stream link
Created service-stream link
Created service-stream link
Created service-stream link
Created service-stream link
Created consumer-service link
Created consumer-service link
Created consumer-service link
Created consumer-service link
Created consumer-service link
Created consumer-service link

✅ Database seeded successfully!
Created:
  - 6 microservices
  - 5 NATS streams
  - 6 consumers
  - 5 service-stream links
  - 6 consumer-service links
```

### Viewing the Data

After seeding, view your topology:

1. **Start the API server** (if not running):
   ```bash
   cd api
   go run main.go
   ```

2. **Open the visualization dashboard**:
   ```bash
   cd visualization
   python -m http.server 8000
   ```
   Visit: http://localhost:8000

3. The dashboard will display your seeded topology with services, streams, and consumers.

### Troubleshooting

#### "Failed to connect to MongoDB"
- Ensure MongoDB is running: `mongod` or `docker run -d -p 27017:27017 mongo`
- Check MONGO_URI is correct
- Verify MongoDB is accessible on the specified host:port

#### "No modules found" (Go error)
- Run from the scripts directory
- Ensure Go 1.21+ is installed
- The seed script doesn't require go.mod in scripts/

#### Data not appearing in visualization
- Verify API server is running on port 8080
- Check browser console for CORS errors
- Visit http://localhost:8080/api/v1/microservices to verify API has data

### Customizing Mock Data

To modify the seeded data:

1. Edit `seed-mongodb.go`
2. Modify the slice definitions (microservices, streams, consumers, etc.)
3. Re-run the script

Example customizations:
- Add more microservices
- Change service names and namespaces
- Add more stream subjects
- Create different link patterns

### Clean Up

To remove all seeded data without running the seed script again:

```bash
# Using MongoDB CLI
mongosh helmjet-atlas
db.microservices.deleteMany({})
db.nats_streams.deleteMany({})
db.nats_consumers.deleteMany({})
db.service_stream_links.deleteMany({})
db.consumer_service_links.deleteMany({})
```

## Future Scripts

Planned utility scripts:
- `backup-mongodb.sh` - Backup data to JSON
- `restore-mongodb.sh` - Restore from backup
- `generate-load.go` - Generate realistic load and metrics
- `cleanup.sh` - Remove all data and reset collections
