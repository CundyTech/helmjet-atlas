document.addEventListener('DOMContentLoaded', () => {
  const width = 900, height = 600;
  const svg = d3.select('#viz').append('svg')
    .attr('width', width)
    .attr('height', height);

  // initial data
  let nodes = [
    { id: 'default/orders-api', name: 'orders-api' },
    { id: 'default/payments-api', name: 'payments-api' },
    { id: 'kube/system', name: 'kube-svc' }
  ];

  let links = [
    { id: 'l1', source: 'default/orders-api', target: 'default/payments-api' },
    { id: 'l2', source: 'default/payments-api', target: 'kube/system' }
  ];

  // map string endpoints to node objects
  function mapLinksToNodes() {
    const nodeById = new Map(nodes.map(n => [n.id, n]));
    links.forEach(l => {
      l.source = typeof l.source === 'string' ? nodeById.get(l.source) : l.source;
      l.target = typeof l.target === 'string' ? nodeById.get(l.target) : l.target;
    });
  }

  // create simulation
  const simulation = d3.forceSimulation()
    .force('link', d3.forceLink().id(d => d.id).distance(120))
    .force('charge', d3.forceManyBody().strength(-300))
    .force('center', d3.forceCenter(width / 2, height / 2));

  // initial render
  function update() {
    mapLinksToNodes();

    const link = svg.selectAll('.link')
      .data(links, d => d.id || `${d.source.id}-${d.target.id}`)
      .join('line')
      .attr('class', 'link');

    const node = svg.selectAll('.node')
      .data(nodes, d => d.id)
      .join(
        enter => {
          const g = enter.append('g').attr('class', 'node');
          g.append('circle').attr('r', 10).attr('fill', '#1f77b4');
          g.append('text').attr('class', 'node-label').attr('x', 12).attr('y', 4).text(d => d.name);
          return g;
        },
        update => update,
        exit => exit.remove()
      );

    // drag behaviour
    node.call(d3.drag()
      .on('start', (event, d) => {
        if (!event.active) simulation.alphaTarget(0.3).restart();
        d.fx = d.x; d.fy = d.y;
      })
      .on('drag', (event, d) => { d.fx = event.x; d.fy = event.y; })
      .on('end', (event, d) => {
        if (!event.active) simulation.alphaTarget(0);
        d.fx = null; d.fy = null;
      })
    );

    simulation.nodes(nodes);
    simulation.force('link').links(links);
    simulation.alpha(1).restart();

    simulation.on('tick', () => {
      svg.selectAll('.link')
        .attr('x1', d => d.source.x)
        .attr('y1', d => d.source.y)
        .attr('x2', d => d.target.x)
        .attr('y2', d => d.target.y);

      svg.selectAll('.node')
        .attr('transform', d => `translate(${d.x},${d.y})`);
    });
  }

  // interactive helpers
  let counter = 0;
  document.getElementById('add').addEventListener('click', () => {
    counter += 1;
    const nid = `default/new-${counter}`;
    nodes.push({ id: nid, name: `new-${counter}` });
    // link new node to orders-api
    links.push({ id: `nl-${counter}`, source: 'default/orders-api', target: nid });
    update();
  });

  document.getElementById('remove').addEventListener('click', () => {
    if (nodes.length <= 1) return;
    const rem = nodes.pop();
    // remove links referencing the removed node
    links = links.filter(l => l.source.id !== rem.id && l.target.id !== rem.id);
    update();
  });

  // initial call
  update();
});
