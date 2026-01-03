# Helmjet Atlas Visualization

A modern, interactive web-based topology visualization dashboard for monitoring and visualizing the relationships between microservices, NATS streams, and consumers in the Helmjet Atlas system.

## Features

- **Interactive Network Topology Graph** - Visualize all services, streams, and consumers as an interactive D3.js force-directed graph
- **Real-time Data** - Auto-refreshes every 30 seconds to show current system state
- **Responsive Design** - Modern dark theme UI with smooth animations
- **Sidebar Statistics** - Quick overview of services, streams, consumers, and their relationships
- **Entity Details** - Click on any node to view detailed information
- **Zoom & Pan** - Intuitive navigation with mouse wheel zoom and drag panning
- **Smart Layout** - Force-directed simulation automatically organizes nodes for clarity

## Getting Started

### Quick Start

Simply open `index.html` in a web browser. The dashboard will automatically connect to the API at `localhost:8080`.

```bash
# Option 1: Open directly in browser
open index.html

# Option 2: Use a local web server (recommended)
python -m http.server 8000
# Then visit http://localhost:8000
```

### Configuration

The visualization connects to the Helmjet Atlas API. To change the API endpoint:

1. Look for the "API Server" input field in the left sidebar
2. Enter your API host and port (e.g., `api.example.com:8080`)
3. Data will refresh automatically

The app connects to these API endpoints:
- `/api/v1/microservices` - List of microservices
- `/api/v1/streams` - List of NATS streams
- `/api/v1/consumers` - List of NATS consumers
- `/api/v1/links/service-stream` - Service-to-stream relationships
- `/api/v1/links/consumer-service` - Consumer-to-service relationships

## Usage

### Navigation

- **Drag nodes** - Click and drag any node to reposition it
- **Zoom** - Use mouse wheel to zoom in/out
- **Pan** - Click and drag the background to pan around
- **Fit to View** - Click the "Fit to View" button to auto-zoom to all nodes
- **Reset Layout** - Click the "Reset Layout" button to reset zoom and pan

### Understanding the Visualization

#### Node Types

- **Blue Circles** - Microservices
- **Red Circles** - NATS Streams
- **Green Circles** - NATS Consumers

#### Links

- Connections between nodes represent relationships
- **Service → Stream** - Service publishes/consumes from stream
- **Consumer → Service** - Consumer processes messages for service

### Sidebar

The left sidebar shows:

- **Statistics** - Count of each entity type and total links
- **Microservices** - All registered microservices (grouped by namespace)
- **NATS Streams** - All available streams (grouped by cluster)
- **Consumers** - All message consumers (grouped by stream)

Click any item to highlight it in the graph and see its details.

### Info Panel

When you select a node, an info panel appears in the top-right showing:

- Entity type
- Key attributes (namespace, cluster, image, etc.)
- Status information
- Configuration details

## Architecture

The visualization is a static HTML/JavaScript application that:

1. Fetches data from the Helmjet Atlas REST API
2. Constructs a force-directed graph using D3.js
3. Renders interactive nodes and links
4. Provides filtering and selection capabilities
5. Auto-refreshes every 30 seconds

No build process or server needed - just open `index.html` in any modern browser.

## Browser Requirements

- Modern browsers with ES6 support
- D3.js v7 (loaded from CDN)
- CORS must be enabled on the API server (or use CORS proxy)

## Styling

The dashboard uses a modern dark theme optimized for long-term viewing. Key colors:

- **Cyan (#00d4ff)** - Primary accent, services, and UI elements
- **Red (#ff6b6b)** - NATS Streams
- **Green (#51cf66)** - Consumers
- **Dark background** - Reduces eye strain during monitoring

## Troubleshooting

### "Failed to load topology data" error

- Check that the API server is running on the configured address
- Verify the API server has CORS enabled
- Check browser console for detailed error messages

### Nodes not appearing

- Ensure there's data in the system (services, streams, or consumers created)
- Check the API endpoints are responding with data
- Verify the browser console shows no JavaScript errors

### Slow performance with many nodes

- The visualization automatically clusters nodes to improve performance
- Reduce the number of displayed entities or increase the browser's hardware acceleration

## Development

To customize the visualization:

1. **Colors** - Edit the `colors` object in `topology.js`
2. **Node sizes** - Modify the `nodeRadius` object
3. **Force simulation** - Adjust `forceStrength` and `linkDistance` for graph layout
4. **Refresh rate** - Change the `setInterval` value at the bottom of `topology.js`
5. **Styles** - Modify the CSS in `index.html` or create a custom stylesheet

## License

Part of the Helmjet Atlas project.
