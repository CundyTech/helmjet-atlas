// Global state
let topologyData = {
    services: [],
    streams: [],
    consumers: [],
};

let simulation = null;
let svg = null;
let g = null;
let zoom = null;
let selectedNode = null;
let autoRefreshInterval = null;
let autoRefreshEnabled = false;
// Namespace filter state (set of selected namespaces)
let selectedNamespaces = new Set();

// Configuration
const CONFIG = {
    apiBase: 'http://localhost:8080/api/v1',
    nodeRadius: {
        service: 20,
        stream: 30,
        consumer: 30
    },
    colors: {
        service: '#00d4ff',
        stream: '#ff6b6b',
        consumer: '#51cf66'
    },
    forceStrength: -700, // Stronger repulsion for groups
    linkDistance: 400,    // Increase distance for groups
    autoRefreshInterval: 30000 // 30 seconds
};

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
    initializeTheme();
    initializeCanvas();
    loadData();
});

function setupEventListeners() {
    document.getElementById('refreshBtn').addEventListener('click', refreshAndSync);
    document.getElementById('autoRefreshBtn').addEventListener('click', toggleAutoRefresh);
    document.getElementById('fitBtn').addEventListener('click', fitToView);
    document.getElementById('resetBtn').addEventListener('click', resetLayout);
    document.getElementById('apiUrl').addEventListener('change', updateApiUrl);
    document.getElementById('themeToggle').addEventListener('click', toggleTheme);
    // Upload handlers
    const k8sBtn = document.getElementById('k8sUploadBtn');
    if (k8sBtn) k8sBtn.addEventListener('click', uploadK8sFile);
    const natsBtn = document.getElementById('natsUploadBtn');
    if (natsBtn) natsBtn.addEventListener('click', uploadNATSFile);

    // Show selected filename
    const k8sInput = document.getElementById('k8sUploadFile');
    const k8sName = document.getElementById('k8sFileName');
    if (k8sInput && k8sName) k8sInput.addEventListener('change', (e) => {
        k8sName.textContent = k8sInput.files && k8sInput.files.length ? k8sInput.files[0].name : 'No file chosen';
    });
    const natsInput = document.getElementById('natsUploadFile');
    const natsName = document.getElementById('natsFileName');
    if (natsInput && natsName) natsInput.addEventListener('change', (e) => {
        natsName.textContent = natsInput.files && natsInput.files.length ? natsInput.files[0].name : 'No file chosen';
    });
}

async function uploadK8sFile() {
    const input = document.getElementById('k8sUploadFile');
    if (!input || !input.files || input.files.length === 0) return alert('Select a JSON file to upload');
    const file = input.files[0];
    const fd = new FormData();
    fd.append('file', file);
    try {
        const resp = await fetch(`${CONFIG.apiBase}/integrations/k8s/upload`, { method: 'POST', body: fd });
        if (!resp.ok) {
            const txt = await resp.text();
            return alert('Upload failed: ' + txt);
        }
        const j = await resp.json();
        alert('K8s upload processed: ' + (j.processed || 0) + ' services');
        await loadData();
    } catch (e) {
        console.warn('K8s upload error', e);
        alert('K8s upload failed');
    }
}

async function uploadNATSFile() {
    const input = document.getElementById('natsUploadFile');
    if (!input || !input.files || input.files.length === 0) return alert('Select a JSON file to upload');
    const file = input.files[0];
    const fd = new FormData();
    fd.append('file', file);
    try {
        const resp = await fetch(`${CONFIG.apiBase}/integrations/nats/upload`, { method: 'POST', body: fd });
        if (!resp.ok) {
            const txt = await resp.text();
            return alert('Upload failed: ' + txt);
        }
        const j = await resp.json();
        alert('NATS upload processed: ' + (j.streams || 0) + ' streams, ' + (j.consumers || 0) + ' consumers');
        await loadData();
    } catch (e) {
        console.warn('NATS upload error', e);
        alert('NATS upload failed');
    }
}

function initializeTheme() {
    // Check localStorage for saved theme preference
    const savedTheme = localStorage.getItem('helmjet-theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

    // Use saved theme, or default to system preference
    const isDarkMode = savedTheme ? savedTheme === 'dark' : prefersDark;

    if (!isDarkMode) {
        document.body.classList.add('light-mode');
        updateThemeButton(false);
    } else {
        updateThemeButton(true);
    }
}

function toggleTheme() {
    const isLightMode = document.body.classList.toggle('light-mode');
    const isDarkMode = !isLightMode;

    // Save preference to localStorage
    localStorage.setItem('helmjet-theme', isDarkMode ? 'dark' : 'light');

    // Update button text
    updateThemeButton(isDarkMode);
}

function updateThemeButton(isDarkMode) {
    const btn = document.getElementById('themeToggle');
    btn.textContent = isDarkMode ? '☀️ Light Mode' : '🌙 Dark Mode';
}

function toggleAutoRefresh() {
    autoRefreshEnabled = !autoRefreshEnabled;
    const btn = document.getElementById('autoRefreshBtn');
    const indicator = document.querySelector('.refresh-indicator');

    if (autoRefreshEnabled) {
        // Start auto-refresh
        autoRefreshInterval = setInterval(loadData, CONFIG.autoRefreshInterval);
        btn.textContent = '⏸️ Auto-Refresh: ON';
        btn.style.background = 'linear-gradient(135deg, #51cf66 0%, #40a050 100%)';
        if (indicator) {
            indicator.style.background = '#51cf66';
            indicator.style.animation = 'pulse 1s infinite';
        }
    } else {
        // Stop auto-refresh
        if (autoRefreshInterval) {
            clearInterval(autoRefreshInterval);
            autoRefreshInterval = null;
        }
        btn.textContent = '▶️ Auto-Refresh: OFF';
        btn.style.background = 'linear-gradient(135deg, #00d4ff 0%, #0099cc 100%)';
        if (indicator) {
            indicator.style.background = '#808080';
            indicator.style.animation = 'none';
        }
    }
}

function updateApiUrl() {
    const url = document.getElementById('apiUrl').value;
    CONFIG.apiBase = `http://${url}/api/v1`;
}

function initializeCanvas() {
    const canvas = document.getElementById('canvas');
    canvas.innerHTML = '<svg></svg>';

    svg = d3.select('svg');
    g = svg.append('g');

    const width = canvas.clientWidth;
    const height = canvas.clientHeight;

    // Add zoom behavior
    zoom = d3.zoom()
        .on('zoom', (event) => {
            g.attr('transform', event.transform);
        });

    svg.call(zoom);

    // Initialize simulation
    simulation = d3.forceSimulation()
        .force('link', d3.forceLink().id(d => d.id).distance(CONFIG.linkDistance))
        .force('charge', d3.forceManyBody().strength(CONFIG.forceStrength))
        .force('center', d3.forceCenter(width / 2, height / 2))
        .force('x', d3.forceX(width / 2).strength(0.1)) // Gentle centering X
        .force('y', d3.forceY(height / 2).strength(0.1)) // Gentle centering Y
        .force('collision', d3.forceCollide(d => (d.width || 50) / 1.5));

    // Drag behavior is defined inline when creating node elements below.

    simulation.on('tick', () => {
        // Update links with curved paths
        g.selectAll('.link').attr('d', d => {
            const start = getAnchorPos(d.source, false); // isTarget=false -> Right
            const end = getAnchorPos(d.target, true);    // isTarget=true  -> Left

            // Simple Bezier curve:
            // Control points are midway horizontally but maintain connection y
            const dx = end.x - start.x;
            const dy = end.y - start.y;
            const dr = Math.sqrt(dx * dx + dy * dy); // Distance

            // For left-to-right flow, we might want control points to extend horizontally
            // Let's try a standard sigmoid-like curve (M S C C1 C2 E)
            // Use 50% of the horizontal distance for control point offsets
            const offset = Math.abs(dx) * 0.5;

            // Or use a fixed offset? fixed often looks cleaner for diagrams if items overlap
            // Let's use an offset proportional to distance but clamped.
            const cpOffset = Math.max(50, Math.abs(dx) / 2);

            return `M${start.x},${start.y} 
                    C${start.x + cpOffset},${start.y} 
                     ${end.x - cpOffset},${end.y} 
                     ${end.x},${end.y}`;
        });

        // Update nodes groups
        g.selectAll('.node')
            .attr('transform', d => `translate(${d.x},${d.y})`);
    });
}

// Global helper to find node position
// isTarget: true = anchor on LEFT side, false = anchor on RIGHT side
function getAnchorPos(nodeOrId, isTarget = true) {
    const id = typeof nodeOrId === 'object' ? nodeOrId.id : nodeOrId;

    // Check if it's a top-level node (Service)
    const service = topologyData.services.find(s => s.id === id);
    if (service && service.x !== undefined) {
        // If target, left side. If source, right side.
        const xOffset = isTarget ? -(service.width / 2) : (service.width / 2);
        return { x: service.x + xOffset, y: service.y };
    }

    // Check if it's a child node (Stream/Consumer)
    const parent = topologyData.services.find(s =>
        (s.children || []).some(c => c.id === id)
    );

    if (parent && parent.x !== undefined) {
        const child = parent.children.find(c => c.id === id);
        // Position arrows at the edge of the service container (external to the box)
        // Links connect to arrows at the container boundary
        const xOffset = isTarget ? -(parent.width / 2) : (parent.width / 2);
        return {
            x: parent.x + xOffset,
            y: parent.y + child.relY
        };
    }

    // Fallback
    if (nodeOrId.x !== undefined) return { x: nodeOrId.x, y: nodeOrId.y };

    return { x: 0, y: 0 };
}

async function loadData() {
    try {
        showLoading(true);

        // Load all data in parallel
        const [services, streams, consumers] = await Promise.all([
            fetchData('/microservices'),
            fetchData('/streams'),
            fetchData('/consumers')
        ]);

        topologyData = {
            services: services || [],
            streams: streams || [],
            consumers: consumers || []
        };

        updateStatistics();
        updateSidebar();
        populateNamespaceSelect();
        showLoading(false);
        updateVisualization();
    } catch (error) {
        console.error('Error loading data:', error);
        showError('Failed to load topology data. Make sure the API is running.');
        showLoading(false);
    }
}

async function fetchData(endpoint) {
    try {
        const response = await fetch(`${CONFIG.apiBase}${endpoint}`);
        if (!response.ok) {
            return [];
        }
        return await response.json();
    } catch (error) {
        console.warn(`Failed to fetch ${endpoint}:`, error);
        return [];
    }
}

function updateStatistics() {
    document.getElementById('serviceCount').textContent = topologyData.services.length;
    document.getElementById('streamCount').textContent = topologyData.streams.length;
    document.getElementById('consumerCount').textContent = topologyData.consumers.length;
}

function updateSidebar() {
    // Update services list
    const servicesList = document.getElementById('servicesList');
    servicesList.innerHTML = topologyData.services.map(s =>
        `<div class="item" data-id="${s.id}" data-type="service">${s.name}<br><span class="item-count">${s.namespace || 'default'}</span></div>`
    ).join('');
    // Update streams list
    const streamsList = document.getElementById('streamsList');
    streamsList.innerHTML = topologyData.streams.map(s =>
        `<div class="item" data-id="${s.id}" data-type="stream">${s.name}<br><span class="item-count">${s.cluster || 'N/A'}</span></div>`
    ).join('');

    // Update consumers list
    const consumersList = document.getElementById('consumersList');
    consumersList.innerHTML = topologyData.consumers.map(c =>
        `<div class="item" data-id="${c.id}" data-type="consumer">${c.name}<br><span class="item-count">${c.streamName}</span></div>`
    ).join('');

    // Event delegation: handle clicks anywhere inside the lists (including inner spans)
    const delegateClick = (e) => {
        const item = e.target.closest('.item');
        if (!item) return;
        const id = item.dataset.id;
        const type = item.dataset.type;
        if (id && type) selectNode(id, type);
    };

    servicesList.onclick = delegateClick;
    streamsList.onclick = delegateClick;
    consumersList.onclick = delegateClick;
}

function populateNamespaceSelect() {
    try {
        const menu = document.getElementById('namespaceDropdownMenu');
        const toggle = document.getElementById('namespaceToggle');
        if (!menu || !toggle) return;

        // Build checkbox list
        const namespaces = Array.from(new Set(topologyData.services.map(s => (s.namespace || 'default')))).sort();
        menu.innerHTML = namespaces.map(ns => `
            <label style="display:flex;align-items:center;margin:4px 0;">
                <input type="checkbox" class="namespace-checkbox" value="${ns}" style="margin-right:8px;">
                <span class="namespace-label">${ns}</span>
            </label>
        `).join('');

        // Toggle dropdown visibility
        toggle.onclick = (e) => {
            e.stopPropagation();
            const visible = menu.style.display === 'block';
            menu.style.display = visible ? 'none' : 'block';
        };

        // Close menu when clicking outside
        document.addEventListener('click', (e) => {
            if (!menu.contains(e.target) && e.target !== toggle) menu.style.display = 'none';
        });

        // Wire checkboxes
        const checkboxes = menu.querySelectorAll('.namespace-checkbox');
        checkboxes.forEach(cb => cb.addEventListener('change', onNamespaceChange));

        // Clear button
        const clearBtn = document.getElementById('clearNamespacesBtn');
        if (clearBtn) {
            clearBtn.onclick = () => {
                menu.querySelectorAll('.namespace-checkbox').forEach(c => c.checked = false);
                onNamespaceChange();
            };
        }
    } catch (e) {
        console.warn('populateNamespaceSelect failed', e);
    }
}

function onNamespaceChange() {
    const menu = document.getElementById('namespaceDropdownMenu');
    if (!menu) return;
    const checked = Array.from(menu.querySelectorAll('.namespace-checkbox')).filter(c => c.checked).map(c => c.value);
    selectedNamespaces = new Set(checked);
    updateVisualization();
}

// Debounced trigger to avoid firing many sync requests rapidly
// (disabled) Namespace-change auto-sync removed — syncs happen via Refresh button only

// Refresh button: trigger both NATS and K8s syncs (inlined) then reload data
async function refreshAndSync() {
    const btn = document.getElementById('refreshBtn');
    if (btn) {
        btn.disabled = true;
        var origText = btn.textContent;
        btn.textContent = '🔄 Refreshing & Syncing...';
    }

    try {
        // Build K8s payload from selected namespaces
        const namespaces = Array.from(selectedNamespaces || []);
        const k8sPayload = { namespaces };

        // Fire both requests in parallel
        const [natsResp, k8sResp] = await Promise.all([
            fetch(`${CONFIG.apiBase}/integrations/nats/sync`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({}) }),
            fetch(`${CONFIG.apiBase}/integrations/k8s/sync`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(k8sPayload) })
        ]);

        if (!natsResp.ok) {
            const txt = await natsResp.text();
            console.warn('NATS sync request failed:', natsResp.status, txt);
        } else {
            try { console.log('NATS sync response:', await natsResp.json()); } catch (e) { }
        }

        if (!k8sResp.ok) {
            const txt = await k8sResp.text();
            console.warn('K8s sync request failed:', k8sResp.status, txt);
        } else {
            try { console.log('K8s sync response:', await k8sResp.json()); } catch (e) { }
        }
    } catch (e) {
        console.warn('Error running syncs:', e);
    }

    // Always refresh local data after attempting syncs
    await loadData();

    if (btn) {
        btn.disabled = false;
        btn.textContent = origText || '🔄 Refresh Data';
    }
}

function updateVisualization() {
    const canvas = document.getElementById('canvas');
    if (!canvas) return;

    // Reset children on services
    topologyData.services.forEach(s => s.children = []);

    // 1. Group nodes into Services
    const groupedNodes = [...topologyData.services];

    // Assign Streams to Parent Services
    topologyData.streams.forEach(stream => {
        let assigned = false;
        if (stream.serviceId) {
            const service = groupedNodes.find(s => s.id === stream.serviceId);
            if (service) {
                service.children.push({ ...stream, type: 'stream' });
                assigned = true;
            }
        }
        if (!assigned) stream._isOrphan = true;
    });

    // Assign Consumers to Parent Services
    topologyData.consumers.forEach(consumer => {
        let assigned = false;
        if (consumer.serviceId) {
            const service = groupedNodes.find(s => s.id === consumer.serviceId);
            if (service) {
                service.children.push({ ...consumer, type: 'consumer' });
                assigned = true;
            }
        }
        if (!assigned) consumer._isOrphan = true;
    });

    // Handle Orphans by creating a "System" group
    const orphanStreams = topologyData.streams.filter(s => s._isOrphan);
    const orphanConsumers = topologyData.consumers.filter(c => c._isOrphan);

    if (orphanStreams.length > 0 || orphanConsumers.length > 0) {
        const systemGroup = {
            id: 'system-unassigned',
            name: 'Unassigned',
            type: 'service',
            children: [
                ...orphanStreams.map(s => ({ ...s, type: 'stream' })),
                ...orphanConsumers.map(c => ({ ...c, type: 'consumer' }))
            ],
            isSystem: true
        };
        groupedNodes.push(systemGroup);
    }

    // Calculate layout for each Group (Service)
    groupedNodes.forEach(service => {
        service.type = 'service';

        // Sort children: Streams first, then Consumers
        service.children.sort((a, b) => {
            if (a.type === b.type) return a.name.localeCompare(b.name);
            return a.type === 'stream' ? -1 : 1; // Stream first
        });

        // ERD Layout Constants
        const headerHeight = 40;
        const itemHeight = 30;
        const width = 200;

        // Calculate heights
        service.contentHeight = (service.children.length * itemHeight);
        service.totalHeight = headerHeight + service.contentHeight;
        service.width = width;
        service.height = service.totalHeight;

        // Position children
        const startY = (-service.totalHeight / 2) + headerHeight;

        service.children.forEach((child, index) => {
            child.relX = 0;
            child.relY = startY + (index * itemHeight) + (itemHeight / 2);
            child.width = width;
            child.rowHeight = itemHeight;
        });
    });

    // 2. Prepare Links
    // Build links from Consumer -> Stream (using embedded streamId)
    // We only visualize Stream -> Consumer data flow
    const realLinks = topologyData.consumers
        .filter(c => c.streamId)
        .map(c => ({
            source: c.streamId, // Source is the Stream
            target: c.id,       // Target is the Consumer
            type: 'stream-consumer'
        }));

    // Synthetic links for Simulation (connect Service to Service)
    const syntheticLinks = [];

    // Helper to find parent ID
    const findParentId = (childId) => {
        const service = groupedNodes.find(s => s.id === childId); // Is it a service?
        if (service) return service.id;

        const parent = groupedNodes.find(s => s.children.some(c => c.id === childId));
        return parent ? parent.id : null;
    };

    realLinks.forEach(link => {
        const sourceParentId = findParentId(link.source);
        const targetParentId = findParentId(link.target);

        if (sourceParentId && targetParentId && sourceParentId !== targetParentId) {
            syntheticLinks.push({ source: sourceParentId, target: targetParentId });
        }
    });

    // Debug: groupedNodes/links available for inspection in debugger if needed
    // Apply namespace filtering to children and services (if namespaces selected)
    try {
        // Filter children by namespace when selection is active
        if (selectedNamespaces && selectedNamespaces.size > 0) {
            groupedNodes.forEach(service => {
                service.children = (service.children || []).filter(c => {
                    const childNs = (c.namespace || service.namespace || 'default');
                    return selectedNamespaces.has(childNs);
                });
            });
        }

        // Filter services that should be visible
        const filteredNodes = groupedNodes.filter(service => {
            if (!selectedNamespaces || selectedNamespaces.size === 0) return true;
            if (service.isSystem) {
                return (service.children || []).length > 0;
            }
            const svcNs = service.namespace || 'default';
            return selectedNamespaces.has(svcNs);
        });

        // Build sets of visible ids (services + their children)
        const visibleIds = new Set();
        const visibleServiceIds = new Set();
        filteredNodes.forEach(s => {
            visibleServiceIds.add(s.id);
            visibleIds.add(s.id);
            (s.children || []).forEach(c => visibleIds.add(c.id));
        });

        // Filter real links so both ends are visible
        const filteredRealLinks = realLinks.filter(l => visibleIds.has(l.source) && visibleIds.has(l.target));

        // Filter synthetic links to only visible services
        const filteredSyntheticLinks = syntheticLinks.filter(l => visibleServiceIds.has(l.source) && visibleServiceIds.has(l.target));

        // Update simulation with filtered nodes and links
        simulation.nodes(filteredNodes);
        simulation.force('link').links(filteredSyntheticLinks);
        simulation.alpha(1).restart();

        // Render filtered graph
        renderVisualization(filteredNodes, filteredRealLinks);
    } catch (e) {
        console.warn('Visualization filter/render failed, falling back to full render', e);
        simulation.nodes(groupedNodes);
        simulation.force('link').links(syntheticLinks);
        simulation.alpha(1).restart();
        renderVisualization(groupedNodes, realLinks);
    }
}

function renderVisualization(nodes, links) {
    // 1. Draw Links (Bottom Layer)
    const getId = (n) => (n && typeof n === 'object' ? n.id : n);
    let linkElements = g.selectAll('.link').data(links, d => `${getId(d.source)}-${getId(d.target)}`);
    linkElements.exit().remove();

    linkElements = linkElements
        .enter()
        .append('path') // Changed from line to path
        .attr('class', 'link')
        .attr('fill', 'none')
        .attr('stroke', '#a0a0a0')
        .attr('stroke-width', 1.5)
        .attr('stroke-opacity', 0.4) // Default low opacity
        .attr('marker-end', 'url(#arrow)')
        .merge(linkElements);

    // 2. Draw Service Groups
    let nodeElements = g.selectAll('.node').data(nodes, d => d.id);
    nodeElements.exit().remove();

    const nodeEnter = nodeElements.enter()
        .append('g')
        .attr('class', 'node')
        .call(d3.drag()
            .on('start', (e, d) => {
                if (!e.active) simulation.alphaTarget(0.3).restart();
                d.fx = d.x;
                d.fy = d.y;
            })
            .on('drag', (e, d) => {
                d.fx = e.x;
                d.fy = e.y;
            })
            .on('end', (e, d) => {
                if (!e.active) simulation.alphaTarget(0);
                d.fx = null;
                d.fy = null;
            }));

    // Service Container Box
    nodeEnter.append('rect')
        .attr('class', 'service-box')
        .attr('rx', 0)
        .attr('ry', 0)
        .attr('fill', '#fff')
        .attr('fill-opacity', 0.95)
        .attr('stroke', '#00d4ff')
        .attr('stroke-width', 2)
        .style('filter', 'drop-shadow(0 4px 6px rgba(0,0,0,0.1))')
        .style('cursor', 'pointer')
        .on('click', (e, d) => {
            e.stopPropagation(); // Prevent event bubbling
            selectNode(d.id, 'service');
        });

    // Header Background
    nodeEnter.append('rect')
        .attr('class', 'service-header-bg')
        .attr('fill', '#e6faff')
        .attr('stroke', 'none');

    // Header Separator Line
    nodeEnter.append('line')
        .attr('class', 'header-line')
        .attr('stroke', '#00d4ff')
        .attr('stroke-width', 1);

    // Merge and update dimensions
    const nodeUpdate = nodeEnter.merge(nodeElements);

    nodeUpdate.select('.service-box')
        .attr('width', d => d.width)
        .attr('height', d => d.height)
        .attr('x', d => -d.width / 2)
        .attr('y', d => -d.height / 2);

    nodeUpdate.select('.service-header-bg')
        .attr('width', d => d.width)
        .attr('height', 40)
        .attr('x', d => -d.width / 2)
        .attr('y', d => -d.height / 2);

    nodeUpdate.select('.header-line')
        .attr('x1', d => -d.width / 2)
        .attr('x2', d => d.width / 2)
        .attr('y1', d => -d.height / 2 + 40)
        .attr('y2', d => -d.height / 2 + 40);

    // Service Label (Header)
    nodeEnter.append('text')
        .attr('class', 'service-label')
        .attr('text-anchor', 'middle')
        .attr('font-weight', 'bold')
        .attr('fill', '#005f73')
        .attr('dominant-baseline', 'middle')
        .style('cursor', 'pointer')
        .on('click', (e, d) => {
            e.stopPropagation();
            selectNode(d.id, 'service');
        });

    nodeUpdate.select('.service-label')
        .attr('y', d => -d.height / 2 + 20)
        .text(d => d.name);

    // 3. Draw Children (Nested Nodes)
    // We restart the join for children every time
    nodeUpdate.each(function (service) {
        const group = d3.select(this);

        let children = group.selectAll('.child-node').data(service.children, c => c.id);
        children.exit().remove();

        const childEnter = children.enter()
            .append('g')
            .attr('class', 'child-node')
            .on('mouseover', handleMouseOver)
            .on('mouseout', handleMouseOut)
            .on('click', (e, d) => {
                e.preventDefault(); // Stop default
                e.stopPropagation(); // Prevent service click
                selectNode(d.id, d.type);
            });

        // Child Row Separator (Line at TOP of each child)
        childEnter.append('line')
            .attr('class', 'row-separator')
            .attr('stroke', '#00d4ff')
            .attr('stroke-opacity', 0.2)
            .attr('stroke-width', 1)
            .attr('x1', -service.width / 2)
            .attr('x2', service.width / 2)
            .attr('y1', d => -d.rowHeight / 2)
            .attr('y2', d => -d.rowHeight / 2);

        // Child Hover Background
        childEnter.append('rect')
            .attr('width', service.width)
            .attr('height', d => d.rowHeight)
            .attr('x', -service.width / 2)
            .attr('y', d => -d.rowHeight / 2)
            .attr('fill', '#00d4ff')
            .attr('fill-opacity', 0)
            .attr('stroke', 'none')
            .on('mouseover', function () { d3.select(this).attr('fill-opacity', 0.1); })
            .on('mouseout', function () { d3.select(this).attr('fill-opacity', 0); });

        // Small Icon on the left
        childEnter.each(function (d) {
            const el = d3.select(this);
            // Center of the icon - Left aligned
            const centerX = -85;

            // color by healthStatus when available
            let iconColor = CONFIG.colors[d.type];
            const hs = d.healthStatus || d.HealthStatus;
            if (hs) {
                if (hs === 'Critical') iconColor = '#ef4444';
                else if (hs === 'Warning') iconColor = '#f59e0b';
                else if (hs === 'Healthy') iconColor = '#10b981';
            }

            el.append('text')
                .attr('x', centerX)
                .attr('y', 5) // Vertically centered
                .attr('text-anchor', 'middle')
                .attr('font-size', '14px')
                .attr('font-weight', '900')
                .attr('font-family', 'sans-serif')
                .attr('fill', iconColor)
                .text(d.type === 'stream' ? 'S' : 'C');
        });

        // Child Label - Left aligned next to icon
        childEnter.append('text')
            .attr('x', -70)
            .attr('y', 5) // Centered vertically
            .attr('text-anchor', 'start')
            .attr('font-size', '12px')
            .attr('font-family', 'ui-monospace, monospace')
            .attr('fill', '#333')
            .text(d => d.name.length > 25 ? d.name.substring(0, 23) + '..' : d.name);

        // Update Children positions
        children.merge(childEnter)
            .attr('transform', d => `translate(${d.relX}, ${d.relY})`);
    });
}

function selectNode(nodeId, nodeType) {
    selectedNode = { id: nodeId, type: nodeType };

    // Simply highlight the clicked node (child or parent)
    d3.selectAll('rect, polygon').attr('opacity', 0.5); // reduced opacity for non-selected

    // Reset service boxes (keep them visible) and selected item
    d3.selectAll('.service-box').attr('opacity', 1).attr('stroke', '#00d4ff');

    // Highlight selected item fully
    // We need to find the specific element for the selected node
    // This simple logic might need improvement but works for visual feedback
    // Ideally we would use classes for selection state

    // Restore opacity for visual feedback - simplified for now
    d3.selectAll('rect, polygon').attr('opacity', 1);

    // Update sidebar selection
    document.querySelectorAll('.item').forEach(item => {
        item.classList.toggle('selected', item.dataset.id === nodeId);
    });

    // Show info panel
    // Find node in flat lists
    const node = [...topologyData.services, ...topologyData.streams, ...topologyData.consumers]
        .find(n => n.id === nodeId);

    if (node) {
        showInfoPanel(node, nodeType);
    }
    // Center/focus the selected node in the SVG (if present)
    try {
        focusOnNode(nodeId);
    } catch (e) {
        // ignore if svg/zoom not initialized yet
    }
}

// Smoothly center and zoom on a node by id (no-op if node not found)
function focusOnNode(nodeId) {
    if (!svg || !zoom || !simulation) return;
    const nodes = simulation.nodes();
    if (!nodes || nodes.length === 0) return;
    // Try to find a top-level node first
    let targetX, targetY;
    const top = nodes.find(n => n.id === nodeId);
    if (top && top.x !== undefined && top.y !== undefined) {
        targetX = top.x;
        targetY = top.y;
    } else {
        // Otherwise locate the child (stream/consumer) inside a parent service
        let found = false;
        for (const parent of nodes) {
            if (!parent.children) continue;
            const child = parent.children.find(c => c.id === nodeId);
            if (child && parent.x !== undefined && parent.y !== undefined) {
                // child.relX/relY are layout offsets relative to parent
                targetX = parent.x + (child.relX || 0);
                targetY = parent.y + (child.relY || 0);
                found = true;
                break;
            }
        }
        if (!found) return; // node not present in simulation yet
    }

    const canvas = document.getElementById('canvas');
    if (!canvas) return;
    const width = canvas.clientWidth;
    const height = canvas.clientHeight;

    // Keep current scale
    const current = d3.zoomTransform(svg.node());
    const scale = current.k || 1;

    // Compute translation so target will be centered
    const tx = width / 2 - targetX * scale;
    const ty = height / 2 - targetY * scale;

    svg.transition()
        .duration(700)
        .call(zoom.transform, d3.zoomIdentity.translate(tx, ty).scale(scale));
}

// Helper function to format bytes in human-readable format
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

// Calculate stream health based on various metrics
function calculateStreamHealth(stream, consumers) {
    const issues = [];
    let healthScore = 100;

    // Check replicas
    if (!stream.replicas || stream.replicas < 1) {
        issues.push('No replicas configured');
        healthScore -= 30;
    } else if (stream.replicas < 3) {
        issues.push('Low replica count (recommended: 3+)');
        healthScore -= 15;
    }

    // Check consumers
    if (consumers.length === 0) {
        issues.push('No active consumers');
        healthScore -= 25;
    }

    // Check message activity
    if (stream.state) {
        if (stream.state.messages === 0) {
            issues.push('No messages in stream');
            healthScore -= 20;
        }
    } else {
        issues.push('No state information available');
        healthScore -= 30;
    }

    // Determine status and color
    let status, color;
    if (healthScore >= 80) {
        status = 'Healthy';
        color = '#10b981'; // green
    } else if (healthScore >= 50) {
        status = 'Warning';
        color = '#f59e0b'; // orange
    } else {
        status = 'Critical';
        color = '#ef4444'; // red
    }

    return {
        status,
        color,
        percentage: Math.max(0, Math.round(healthScore)),
        issues
    };
}

function showInfoPanel(node, nodeType) {
    const panel = document.getElementById('infoPanel');
    let content = `<h3>${node.name}</h3>`;

    if (nodeType === 'service') {
        // Find related streams (streams with this service's ID)
        const relatedStreams = topologyData.streams
            .filter(s => s.serviceId === node.id);

        // Find related consumers (consumers with this service's ID)
        const relatedConsumers = topologyData.consumers
            .filter(c => c.serviceId === node.id);

        content += `
            <div class="info-row"><span class="info-label">Type:</span><span class="info-value">Microservice</span></div>
            <div class="info-row"><span class="info-label">Namespace:</span><span class="info-value">${node.namespace || 'default'}</span></div>
            <div class="info-row"><span class="info-label">Cluster:</span><span class="info-value">${node.cluster || 'N/A'}</span></div>
            <div class="info-row"><span class="info-label">Image:</span><span class="info-value">${node.image || 'N/A'}</span></div>
            <div class="info-row"><span class="info-label">Replicas:</span><span class="info-value">${node.replicas || 0}</span></div>
            <div class="info-row"><span class="info-label">Status:</span><span class="info-value">${node.status || 'Unknown'}</span></div>
            
            <div class="info-row"><span class="info-label">Streams:</span><span class="info-value">${relatedStreams.length ? relatedStreams.map(s => s.name).join(', ') : 'None'}</span></div>
            <div class="info-row"><span class="info-label">Consumers:</span><span class="info-value">${relatedConsumers.length ? relatedConsumers.map(c => c.name).join(', ') : 'None'}</span></div>
        `;
    } else if (nodeType === 'stream') {
        // Find parent service (stream has serviceId)
        const parentService = node.serviceId
            ? topologyData.services.find(s => s.id === node.serviceId)
            : null;

        // Find consumers consuming from this stream
        const relatedConsumers = topologyData.consumers
            .filter(c => c.streamId === node.id);

        // If backend provided health fields, use them; otherwise calculate locally
        const backendHealthStatus = node.healthStatus || node.HealthStatus;
        const backendUsage = node.usagePct || node.UsagePct;
        const backendWarnings = node.warnings || node.Warnings || [];

        let health = null;
        if (!backendHealthStatus) {
            // Calculate health for this stream as fallback
            health = calculateStreamHealth(node, relatedConsumers);
        }

        // Build subjects list
        const subjectsList = (node.subjects || []).length > 0
            ? node.subjects.map(s => `<div style="margin-left: 10px;">• ${s}</div>`).join('')
            : '<div style="margin-left: 10px;">None</div>';

        // Build consumers list with their services
        let consumersList = '';
        if (relatedConsumers.length > 0) {
            consumersList = relatedConsumers.map(c => {
                const consumerService = c.serviceId
                    ? topologyData.services.find(s => s.id === c.serviceId)
                    : null;
                const serviceName = consumerService ? consumerService.name : 'Unknown';
                return `<div style="margin-left: 10px;">• ${c.name} <span style="color: #808080;">(${serviceName})</span></div>`;
            }).join('');
        } else {
            consumersList = '<div style="margin-left: 10px;">None</div>';
        }

        // Health display: prefer backend fields when present
        const displayStatus = backendHealthStatus || (health ? health.status : 'Unknown');
        const displayPct = backendUsage !== undefined ? backendUsage : (health ? health.percentage : null);
        const statusColor = displayStatus === 'Critical' ? '#ef4444' : (displayStatus === 'Warning' ? '#f59e0b' : '#10b981');

        content += `
            <div class="info-row"><span class="info-label">Type:</span><span class="info-value">NATS Stream</span></div>
            <div class="info-row"><span class="info-label">Cluster:</span><span class="info-value">${node.cluster || 'N/A'}</span></div>
            <div class="info-row"><span class="info-label">Replicas:</span><span class="info-value">${node.replicas || 0}</span></div>
            <div class="info-row"><span class="info-label">Storage:</span><span class="info-value">${node.storage || 'N/A'}</span></div>

            <div class="info-row"><span class="info-label">Health:</span><span class="info-value"><span style="display:inline-block;width:10px;height:10px;background:${statusColor};border-radius:50%;margin-right:6px;"></span>${displayStatus}${displayPct ? ' (' + displayPct.toFixed(1) + '%)' : ''}</span></div>

            <div class="info-row"><span class="info-label">Messages:</span><span class="info-value">${node.state?.messages?.toLocaleString() || '0'}</span></div>
            <div class="info-row"><span class="info-label">Size:</span><span class="info-value">${formatBytes(node.state?.bytes || 0)}</span></div>
            <div class="info-row"><span class="info-label">Seq Range:</span><span class="info-value">${node.state?.firstSeq || 0} - ${node.state?.lastSeq || 0}</span></div>

            <div class="info-row"><span class="info-label">Service:</span><span class="info-value">${parentService ? parentService.name : 'None'}</span></div>

            <div style="margin-top: 10px; padding-top: 10px; border-top: 1px solid #404060;">
                <div style="color: #00d4ff; font-weight: bold; margin-bottom: 5px;">Subjects (${(node.subjects || []).length}):</div>
                ${subjectsList}
            </div>

            <div style="margin-top: 10px; padding-top: 10px; border-top: 1px solid #404060;">
                <div style="color: #00d4ff; font-weight: bold; margin-bottom: 5px;">Consumers (${relatedConsumers.length}):</div>
                ${consumersList}
            </div>
        `;

        // Show backend warnings if present
        const warns = backendWarnings.length ? backendWarnings : (node.warnings || node.Warnings || []);
        if (warns && warns.length > 0) {
            content += `<div style="margin-top:10px;color:#ef4444;"><strong>Warnings:</strong><ul style="margin-left:14px;">${warns.map(w => `<li>${w}</li>`).join('')}</ul></div>`;
        }

        // If there are health issues, list them below
        if (health.issues && health.issues.length > 0) {
            content += `<div style="margin-top:10px;color:#ef4444;"><strong>Health Issues:</strong><ul style="margin-left:14px;">${health.issues.map(i => `<li>${i}</li>`).join('')}</ul></div>`;
        }
    } else if (nodeType === 'consumer') {
        // Find parent service (consumer has serviceId)
        const parentService = node.serviceId
            ? topologyData.services.find(s => s.id === node.serviceId)
            : null;

        // Find source stream (consumer has streamId)
        const sourceStream = node.streamId
            ? topologyData.streams.find(s => s.id === node.streamId)
            : null;

        // Build subjects list for consumer
        const consumerSubjectsList = (node.subjects || []).length > 0
            ? node.subjects.map(s => `<div style="margin-left: 10px;">• ${s}</div>`).join('')
            : '<div style="margin-left: 10px;">None</div>';

        // Format ack wait time (milliseconds to seconds)
        const ackWaitSeconds = node.ackWait ? (node.ackWait / 1000).toFixed(1) + 's' : 'N/A';

        // Format rate limit
        const rateLimitFormatted = node.rateLimit ? formatBytes(node.rateLimit) + '/s' : 'Unlimited';

        content += `
            <div class="info-row"><span class="info-label">Type:</span><span class="info-value">Consumer</span></div>
            <div class="info-row"><span class="info-label">Stream:</span><span class="info-value">${sourceStream ? sourceStream.name : (node.streamName || 'N/A')}</span></div>
            <div class="info-row"><span class="info-label">Cluster:</span><span class="info-value">${node.cluster || 'N/A'}</span></div>
            <div class="info-row"><span class="info-label">Group:</span><span class="info-value">${node.consumerGroup || 'N/A'}</span></div>
            <div class="info-row"><span class="info-label">Status:</span><span class="info-value">${node.status || 'Active'}</span></div>

            <div class="info-row"><span class="info-label">Service:</span><span class="info-value">${parentService ? parentService.name : 'None'}</span></div>
            
            <div style="margin-top: 10px; padding-top: 10px; border-top: 1px solid #404060;">
                <div style="color: #00d4ff; font-weight: bold; margin-bottom: 5px;">Configuration:</div>
                <div class="info-row"><span class="info-label">Delivery Policy:</span><span class="info-value">${node.deliveryPolicy?.type || 'N/A'}</span></div>
                <div class="info-row"><span class="info-label">Ack Policy:</span><span class="info-value">${node.ackPolicy || 'N/A'}</span></div>
                <div class="info-row"><span class="info-label">Ack Wait:</span><span class="info-value">${ackWaitSeconds}</span></div>
                <div class="info-row"><span class="info-label">Max Deliver:</span><span class="info-value">${node.maxDeliver || 'N/A'}</span></div>
                <div class="info-row"><span class="info-label">Rate Limit:</span><span class="info-value">${rateLimitFormatted}</span></div>
            </div>
            
            <div style="margin-top: 10px; padding-top: 10px; border-top: 1px solid #404060;">
                <div style="color: #00d4ff; font-weight: bold; margin-bottom: 5px;">Subjects (${(node.subjects || []).length}):</div>
                ${consumerSubjectsList}
            </div>
        `;
    }

    panel.innerHTML = content;
    panel.classList.add('active');
}

// Focus Mode: Highlight links on hover
function handleMouseOver(event, d) {
    // Only focus on Stream/Consumer nodes, not Service nodes
    //if (d.children !== undefined) return;

    const id = d.id;
    const getId = (n) => (n && typeof n === 'object' ? n.id : n);

    // 1. Dim ALL links/nodes first
    d3.selectAll('.link')
        .filter(function () { return !this.classList.contains('active-link'); })
        .transition().duration(200).style('stroke-opacity', 0.1);

    d3.selectAll('.node').transition().duration(200).style('opacity', 0.5); // Dim service boxes
    d3.selectAll('.child-node').transition().duration(200).style('opacity', 0.5); // Dim child nodes
    d3.selectAll('.child-node text').transition().duration(200).style('opacity', 0.3); // Dim child node text

    // 2. Highlight THIS node (the hovered child node)
    d3.select(event.currentTarget)
        .transition().duration(200)
        .style('opacity', 1)
        .selectAll('text')
        .style('opacity', 1)
        .style('font-weight', 'bold');

    // Stronger hover background for the hovered child node
    d3.select(event.currentTarget).select('rect[fill="#00d4ff"]')
        .attr('fill-opacity', 0.4);

    // 3. Highlight connected links
    const connectedLinks = d3.selectAll('.link').filter(l => {
        const s = getId(l.source);
        const t = getId(l.target);
        return s === id || t === id;
    });

    connectedLinks
        .transition().duration(200)
        .style('stroke-opacity', 1)
        .style('stroke', '#555')
        .style('stroke-width', 2.5);

    // 4. Highlight connected nodes (child nodes and their parent service boxes)
    connectedLinks.each(function (l) {
        const s = getId(l.source);
        const t = getId(l.target);
        const neighborId = s === id ? t : s;

        // Highlight the neighbor child node
        d3.selectAll('.child-node').filter(n => n.id === neighborId)
            .transition().duration(200)
            .style('opacity', 1)
            .selectAll('text')
            .style('opacity', 1)
            .style('font-weight', 'normal'); // Keep neighbor text normal, only hovered is bold

        // Highlight the parent service box of the neighbor
        d3.selectAll('.node').filter(n => n.children && n.children.some(c => c.id === neighborId))
            .transition().duration(200).style('opacity', 1);
    });
}

function handleMouseOut(event, d) {
    // Reset everything
    d3.selectAll('.link').transition().duration(200)
        .style('stroke-opacity', 0.4)
        .style('stroke', '#a0a0a0')
        .style('stroke-width', 1.5);

    d3.selectAll('.node').transition().duration(200).style('opacity', 1);
    d3.selectAll('.child-node text').transition().duration(200).style('opacity', 1);

    // Reset background
    d3.select(event.currentTarget).select('rect[fill="#00d4ff"]')
        .attr('fill-opacity', 0);
}

function fitToView() {
    if (!simulation.nodes().length) return;

    const nodes = simulation.nodes();
    let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;

    nodes.forEach(d => {
        minX = Math.min(minX, d.x);
        maxX = Math.max(maxX, d.x);
        minY = Math.min(minY, d.y);
        maxY = Math.max(maxY, d.y);
    });

    const canvas = document.getElementById('canvas');
    const width = canvas.clientWidth;
    const height = canvas.clientHeight;

    const padding = 50;
    const nodeWidth = maxX - minX + 2 * padding;
    const nodeHeight = maxY - minY + 2 * padding;

    const scale = Math.min(width / nodeWidth, height / nodeHeight, 2);
    const centerX = (width - scale * (minX + maxX)) / 2;
    const centerY = (height - scale * (minY + maxY)) / 2;

    svg.transition()
        .duration(500)
        .call(zoom.transform, d3.zoomIdentity.translate(centerX, centerY).scale(scale));
}

function resetLayout() {
    const canvas = document.getElementById('canvas');
    const width = canvas.clientWidth;
    const height = canvas.clientHeight;

    svg.transition()
        .duration(500)
        .call(zoom.transform, d3.zoomIdentity.translate(0, 0).scale(1));

    simulation.force('center', d3.forceCenter(width / 2, height / 2));
    simulation.alpha(1).restart();
}

function showLoading(show) {
    const canvas = document.getElementById('canvas');
    if (show) {
        canvas.innerHTML = '<div class="loading"><div class="spinner"></div>Loading topology data...</div>';
    } else {
        // Clear the loading message by reinitializing the canvas with SVG
        canvas.innerHTML = '<svg></svg>';
        svg = d3.select('svg');
        g = svg.append('g');

        zoom = d3.zoom()
            .on('zoom', (event) => {
                g.attr('transform', event.transform);
            });

        svg.call(zoom);

        // Restore arrow marker
        svg.append('defs').append('marker')
            .attr('id', 'arrow')
            .attr('viewBox', '0 -5 10 10')
            .attr('refX', 10) // Attach tip exactly to path end
            .attr('refY', 0)
            .attr('markerWidth', 6)
            .attr('markerHeight', 6)
            .attr('orient', 'auto')
            .append('path')
            .attr('d', 'M0,-5L10,0L0,5')
            .attr('fill', '#999');
    }
}

function showError(message) {
    const canvas = document.getElementById('canvas');
    canvas.innerHTML = `<div class="loading" style="color: #ff6b6b;">${message}</div>`;
}

// Auto-refresh every 30 seconds
//setInterval(loadData, 30000);