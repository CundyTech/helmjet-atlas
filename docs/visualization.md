# Visualization / topology (D3) documentation

This document explains how the topology visualization (in [visualization/topology.js](visualization/topology.js)) uses D3 (v7) to render services and their links and shows concise examples for joining/updating nodes and links.

## Data model

- nodes: an array of service objects. Each node should have a stable unique id (string) — e.g. `${namespace}/${name}` — and visual fields such as `name`, `namespace`, `ports`, and coordinates when computed by the force layout.
- links: an array of link objects where each link references source and target by id (string) or by index. Example link: `{ id: "link-1", source: "default/orders-api", target: "default/payments-api" }`.

Before binding `links` to DOM elements, convert `source`/`target` ids into object references that D3's force simulation expects (objects from the `nodes` array). Example:

```js
// nodes: [{ id: 'default/orders-api', ... }, ...]
// links: [{ source: 'default/orders-api', target: 'default/payments-api' }, ...]
const nodeById = new Map(nodes.map(n => [n.id, n]));
links.forEach(l => {
  l.source = nodeById.get(typeof l.source === 'string' ? l.source : l.source.id);
  l.target = nodeById.get(typeof l.target === 'string' ? l.target : l.target.id);
});
```

## The D3 join pattern (enter / update / exit)

Use D3's selection.data(key) with a stable key function (the node `id`) so D3 can match existing DOM elements to data across updates.

Nodes example (group + circle + label):

```js
const nodeSel = svg.selectAll('.node').data(nodes, d => d.id);

const nodeEnter = nodeSel.enter().append('g').attr('class', 'node');
nodeEnter.append('circle').attr('r', 8).attr('class', 'node-circle');
nodeEnter.append('text').attr('class', 'node-label').text(d => d.name);

// Merge enter + update
const nodeMerge = nodeEnter.merge(nodeSel);

// Update common attributes (positions applied in the tick handler)
nodeMerge.attr('data-namespace', d => d.namespace);

// Remove exited nodes
nodeSel.exit().remove();
```

Links example (lines):

```js
const linkSel = svg.selectAll('.link').data(links, d => d.id || `${d.source.id}-${d.target.id}`);

linkSel.enter().append('line').attr('class', 'link');
linkSel.exit().remove();
```

Alternatively, use the `selection.join()` convenience (shorter form):

```js
const link = svg.selectAll('.link')
  .data(links, d => d.id || `${d.source.id}-${d.target.id}`)
  .join('line')
  .attr('class', 'link');

const node = svg.selectAll('.node')
  .data(nodes, d => d.id)
  .join('g')
  .attr('class', 'node');
```

## Force simulation and tick updates

When using `d3.forceSimulation(nodes)`, update the `x`/`y` attributes of the SVG elements each simulation tick:

```js
simulation.on('tick', () => {
  svg.selectAll('.link')
    .attr('x1', d => d.source.x)
    .attr('y1', d => d.source.y)
    .attr('x2', d => d.target.x)
    .attr('y2', d => d.target.y);

  svg.selectAll('.node')
    .attr('transform', d => `translate(${d.x},${d.y})`);
});
```

Note: `d.source` and `d.target` must be object references (see the `nodeById` mapping above) so the simulation sets `x`/`y` on those objects.

## Updating data (example flow)

1. Compute new `nodes` and `links` arrays (e.g., reloaded from the backend).
2. Map link endpoints to node objects using a stable id map.
3. Re-bind `links` and `nodes` with `.data(newArr, key)` and run the `join` or `enter/merge/exit` pattern.
4. Restart or update the force simulation's node/link references and call `simulation.alpha(1).restart()` if you want the layout to re-run.

Example applying steps 2–4:

```js
// 1) update arrays
nodes = fetchedNodes;
links = fetchedLinks;

// 2) remap link endpoints
const nodeById = new Map(nodes.map(n => [n.id, n]));
links.forEach(l => { l.source = nodeById.get(l.source); l.target = nodeById.get(l.target); });

// 3) join
const link = svg.selectAll('.link').data(links, d => d.id).join('line').attr('class','link');
const node = svg.selectAll('.node').data(nodes, d => d.id).join(
  enter => {
    const g = enter.append('g').attr('class','node');
    g.append('circle').attr('r',8).attr('class','node-circle');
    g.append('text').attr('class','node-label').text(d=>d.name);
    return g;
  },
  update => update,
  exit => exit => exit.remove()
);

// 4) update simulation
simulation.nodes(nodes);
simulation.force('link').links(links);
simulation.alpha(1).restart();
```

## Common pitfalls and tips

- Key selection: always provide a stable key function to `.data()` (`d => d.id`) — otherwise nodes will be recreated and transitions will be lost.
- Link endpoint types: ensure `link.source`/`link.target` are object references before starting the simulation.
- Performance: for large graphs, throttle updates and avoid recreating the full SVG tree on frequent refreshes; update attributes instead of full re-creation when possible.
- Labels & interactivity: place text elements in the node group and use pointer events on the group for drag/click handling.

## Where to look in the codebase

- Main visualization code: [visualization/topology.js](visualization/topology.js)
- Static HTML: [visualization/index.html](visualization/index.html)
- Styles for visualization: [visualization/styles.css](visualization/styles.css)

If you'd like, I can add a runnable minimal example page under `visualization/examples/` that demonstrates a simple nodes+links join and drag behavior you can open in the browser. Would you like that? 
