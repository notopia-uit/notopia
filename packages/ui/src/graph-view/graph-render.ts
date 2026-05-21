import { Group as TweenGroup, Tween as Tweened } from '@tweenjs/tween.js';
import {
  SimulationNodeDatum,
  SimulationLinkDatum,
  Simulation,
  forceSimulation,
  forceManyBody,
  forceCenter,
  forceLink,
  forceCollide,
  forceRadial,
  zoomIdentity,
  select,
  drag,
  zoom,
} from 'd3';
import { Text, Graphics, Application, Container, Circle } from 'pixi.js';

import type { D3Config } from './graph';

type GraphicsInfo = {
  color: string;
  gfx: Graphics;
  alpha: number;
  active: boolean;
};

type NodeData = {
  id: string;
  name: string;
  type: 'note' | 'tag';
  weight?: number;
} & SimulationNodeDatum;

type LinkData = {
  source: NodeData;
  target: NodeData;
} & SimulationLinkDatum<NodeData>;

type LinkRenderData = GraphicsInfo & {
  simulationData: LinkData;
};

type NodeRenderData = GraphicsInfo & {
  simulationData: NodeData;
  label: Text;
};

type TweenNode = {
  update: (time: number) => void;
  stop: () => void;
};

const localStorageKey = 'graph-visited';

function getVisited(): Set<string> {
  if (typeof window === 'undefined') return new Set();
  return new Set(JSON.parse(localStorage.getItem(localStorageKey) ?? '[]'));
}

function addToVisited(slug: string) {
  if (typeof window === 'undefined') return;
  const visited = getVisited();
  visited.add(slug);
  localStorage.setItem(localStorageKey, JSON.stringify([...visited]));
}

export function removeAllChildren(element: HTMLElement) {
  while (element.firstChild) {
    element.removeChild(element.firstChild);
  }
}

export async function renderGraph(
  graph: HTMLElement,
  currentSlug: string,
  graphData: any,
  config: Partial<D3Config>
): Promise<() => void> {
  const visited = getVisited();
  removeAllChildren(graph);

  const {
    drag: enableDrag = true,
    zoom: enableZoom = true,
    depth = 1,
    scale = 1.1,
    repelForce = 0.5,
    centerForce = 0.3,
    linkDistance = 30,
    fontSize = 0.6,
    opacityScale = 1,
    removeTags = [],
    showTags = true,
    focusOnHover = false,
    enableRadial = false,
  } = config;

  // Filter nodes based on showTags and removeTags config
  const isTagNode = (node: NodeData) => node.type === 'tag';

  const filteredNodes = graphData.nodes.filter((node: NodeData) => {
    const isTag = isTagNode(node);

    // Filter by showTags setting
    if (isTag && !showTags) {
      return false;
    }

    // Filter by removeTags list
    if (removeTags.length > 0 && isTag) {
      // Extract tag name from node id or use node name
      const tagName = node.id.replace('#', '').split('-')[0] || node.name.toLowerCase();
      if (removeTags.includes(tagName)) {
        return false;
      }
    }

    return true;
  });

  // Create node map for quick lookup
  const nodeMap = new Map(filteredNodes.map((node: NodeData) => [node.id, node]));
  const validNodeIds = new Set(filteredNodes.map((n: NodeData) => n.id));

  const tweens = new Map<string, TweenNode>();

  // Filter links to only include valid nodes
  const validLinks = graphData.links.filter(
    (link: any) => validNodeIds.has(link.source) && validNodeIds.has(link.target)
  );

  // Calculate neighborhood based on depth
  const neighbourhood = new Set<string>();

  // If currentSlug doesn't exist in the filtered nodes, use the first available node
  let startNode = currentSlug;
  if (!validNodeIds.has(currentSlug) && validNodeIds.size > 0) {
    startNode = Array.from(validNodeIds)[0] as string;
  }

  const wl: (string | '__SENTINEL')[] = [startNode, '__SENTINEL'];

  if (depth >= 0) {
    let currentDepth = depth;
    while (currentDepth >= 0 && wl.length > 0) {
      const cur = wl.shift()!;
      if (cur === '__SENTINEL') {
        currentDepth--;
        wl.push('__SENTINEL');
      } else {
        neighbourhood.add(cur);
        const outgoing = validLinks.filter((l: any) => l.source === cur);
        const incoming = validLinks.filter((l: any) => l.target === cur);
        wl.push(...outgoing.map((l: any) => l.target), ...incoming.map((l: any) => l.source));
      }
    }
  } else {
    validNodeIds.forEach((id: any) => neighbourhood.add(id));
  }

  // Build nodes for rendering
  const nodes: NodeData[] = [...neighbourhood]
    .map((id: string) => nodeMap.get(id))
    .filter((node): node is NodeData => node !== undefined);

  // Build links for rendering
  const processedGraphData: { nodes: NodeData[]; links: LinkData[] } = {
    nodes,
    links: validLinks
      .filter((l: any) => neighbourhood.has(l.source) && neighbourhood.has(l.target))
      .map((l: any) => ({
        source: nodes.find((n: NodeData) => n.id === l.source)!,
        target: nodes.find((n: NodeData) => n.id === l.target)!,
      })),
  };

  const width = graph.offsetWidth;
  const height = Math.max(graph.offsetHeight, 250);

  // D3 force simulation
  const simulation: Simulation<NodeData, LinkData> = forceSimulation<NodeData>(
    processedGraphData.nodes
  )
    .force('charge', forceManyBody().strength(-100 * repelForce))
    .force('center', forceCenter().strength(centerForce))
    .force('link', forceLink(processedGraphData.links).distance(linkDistance))
    .force('collide', forceCollide<NodeData>((n) => nodeRadius(n)).iterations(3));

  const radius = (Math.min(width, height) / 2) * 0.8;
  if (enableRadial) simulation.force('radial', forceRadial(radius).strength(0.2));

  // Get CSS variables and convert them to safe hex colors for Pixi.js
  const cssVars = [
    '--secondary',
    '--tertiary',
    '--gray',
    '--light',
    '--lightgray',
    '--dark',
    '--darkgray',
    '--bodyFont',
  ] as const;

  const getDefaultColor = (varName: string): string => {
    const defaults: Record<string, string> = {
      '--secondary': '#4e9f3d',
      '--tertiary': '#8e44ad',
      '--gray': '#8c8c8c',
      '--light': '#ffffff',
      '--lightgray': '#e0e0e0',
      '--dark': '#333333',
      '--darkgray': '#666666',
      '--bodyFont': 'Arial, sans-serif',
    };
    return defaults[varName] || '#000000';
  };

  const convertColorToHex = (colorValue: string): string => {
    // If already a hex color, return it
    if (colorValue.startsWith('#')) {
      return colorValue;
    }

    // If empty, return default
    if (!colorValue.trim()) {
      return '#000000';
    }

    // Try to parse HSL colors (they come as "h s% l%")
    const hslMatch = colorValue.match(/^(\d+\.?\d*)\s+(\d+\.?\d*)%\s+(\d+\.?\d*)%$/);
    if (hslMatch) {
      const [, h, s, l] = hslMatch.map(Number);
      // Convert HSL to hex
      const hslToRgb = (h: number, s: number, l: number) => {
        h = h / 360;
        s = s / 100;
        l = l / 100;

        let r, g, b;
        if (s === 0) {
          r = g = b = l;
        } else {
          const hue2rgb = (p: number, q: number, t: number) => {
            if (t < 0) t += 1;
            if (t > 1) t -= 1;
            if (t < 1 / 6) return p + (q - p) * 6 * t;
            if (t < 1 / 2) return q;
            if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
            return p;
          };

          const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
          const p = 2 * l - q;
          r = hue2rgb(p, q, h + 1 / 3);
          g = hue2rgb(p, q, h);
          b = hue2rgb(p, q, h - 1 / 3);
        }

        const toHex = (x: number) => {
          const hex = Math.round(x * 255).toString(16);
          return hex.length === 1 ? '0' + hex : hex;
        };

        return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
      };
      return hslToRgb(h, s, l);
    }

    // For any other color format (rgb, hsl with parentheses, lab, etc.),
    // use a canvas element to parse it
    try {
      const canvas = document.createElement('canvas');
      canvas.width = canvas.height = 1;
      const ctx = canvas.getContext('2d');
      if (!ctx) return '#000000';

      ctx.fillStyle = colorValue;
      ctx.fillRect(0, 0, 1, 1);

      const imageData = ctx.getImageData(0, 0, 1, 1);
      const [r, g, b] = imageData.data;

      const toHex = (x: number) => {
        const hex = x.toString(16);
        return hex.length === 1 ? '0' + hex : hex;
      };

      return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
    } catch {
      return '#000000';
    }
  };

  const computedStyleMap = cssVars.reduce(
    (acc, key) => {
      const value = getComputedStyle(document.documentElement).getPropertyValue(key).trim();
      const colorValue = value || getDefaultColor(key);
      // Convert all colors to hex format safe for Pixi.js
      acc[key] = key === '--bodyFont' ? colorValue : convertColorToHex(colorValue);
      return acc;
    },
    {} as Record<(typeof cssVars)[number], string>
  );

  const color = (d: NodeData) => {
    const isCurrent = d.id === currentSlug;

    // If this is the current node, highlight it
    if (isCurrent) {
      return computedStyleMap['--secondary'];
    }

    // All tags get the same tag color
    if (d.type === 'tag') {
      return computedStyleMap['--tertiary'];
    }

    // All notes get the same note color
    return computedStyleMap['--gray'];
  };

  function nodeRadius(d: NodeData) {
    const numLinks = processedGraphData.links.filter(
      (l) => l.source.id === d.id || l.target.id === d.id
    ).length;
    return 2 + Math.sqrt(numLinks);
  }

  let hoveredNodeId: string | null = null;
  let hoveredNeighbours: Set<string> = new Set();
  const linkRenderData: LinkRenderData[] = [];
  const nodeRenderData: NodeRenderData[] = [];

  function updateHoverInfo(newHoveredId: string | null) {
    hoveredNodeId = newHoveredId;

    if (newHoveredId === null) {
      hoveredNeighbours = new Set();
      for (const n of nodeRenderData) {
        n.active = false;
      }
      for (const l of linkRenderData) {
        l.active = false;
      }
    } else {
      hoveredNeighbours = new Set();
      for (const l of linkRenderData) {
        const linkData = l.simulationData;
        if (linkData.source.id === newHoveredId || linkData.target.id === newHoveredId) {
          hoveredNeighbours.add(linkData.source.id);
          hoveredNeighbours.add(linkData.target.id);
        }
        l.active = linkData.source.id === newHoveredId || linkData.target.id === newHoveredId;
      }
      for (const n of nodeRenderData) {
        n.active = hoveredNeighbours.has(n.simulationData.id);
      }
    }
  }

  let dragStartTime = 0;
  let dragging = false;

  function renderLinks() {
    tweens.get('link')?.stop();
    const tweenGroup = new TweenGroup();

    for (const l of linkRenderData) {
      let alpha = 1;
      if (hoveredNodeId) {
        alpha = l.active ? 1 : 0.2;
      }
      l.color = l.active ? computedStyleMap['--gray'] : computedStyleMap['--lightgray'];
      tweenGroup.add(new Tweened<LinkRenderData>(l).to({ alpha }, 200));
    }

    tweenGroup.getAll().forEach((tw) => tw.start());
    tweens.set('link', {
      update: tweenGroup.update.bind(tweenGroup),
      stop() {
        tweenGroup.getAll().forEach((tw) => tw.stop());
      },
    });
  }

  function renderLabels() {
    tweens.get('label')?.stop();
    const tweenGroup = new TweenGroup();

    const defaultScale = 1 / scale;
    const activeScale = defaultScale * 1.1;

    for (const n of nodeRenderData) {
      const nodeId = n.simulationData.id;

      if (hoveredNodeId === nodeId) {
        tweenGroup.add(
          new Tweened<Text>(n.label).to(
            {
              alpha: 1,
              scale: { x: activeScale, y: activeScale },
            },
            100
          )
        );
      } else {
        tweenGroup.add(
          new Tweened<Text>(n.label).to(
            {
              alpha: n.label.alpha,
              scale: { x: defaultScale, y: defaultScale },
            },
            100
          )
        );
      }
    }

    tweenGroup.getAll().forEach((tw) => tw.start());
    tweens.set('label', {
      update: tweenGroup.update.bind(tweenGroup),
      stop() {
        tweenGroup.getAll().forEach((tw) => tw.stop());
      },
    });
  }

  function renderNodes() {
    tweens.get('hover')?.stop();
    const tweenGroup = new TweenGroup();

    for (const n of nodeRenderData) {
      let alpha = 1;
      if (hoveredNodeId !== null && focusOnHover) {
        alpha = n.active ? 1 : 0.2;
      }
      tweenGroup.add(new Tweened<Graphics>(n.gfx, tweenGroup).to({ alpha }, 200));
    }

    tweenGroup.getAll().forEach((tw) => tw.start());
    tweens.set('hover', {
      update: tweenGroup.update.bind(tweenGroup),
      stop() {
        tweenGroup.getAll().forEach((tw) => tw.stop());
      },
    });
  }

  function renderPixiFromD3() {
    renderNodes();
    renderLinks();
    renderLabels();
  }

  tweens.forEach((tween) => tween.stop());
  tweens.clear();

  // Initialize Pixi.js
  const app = new Application();
  await app.init({
    width,
    height,
    antialias: true,
    autoStart: false,
    autoDensity: true,
    backgroundAlpha: 0,
    preference: 'webgpu',
    resolution: window.devicePixelRatio,
    eventMode: 'static',
  });
  graph.appendChild(app.canvas);

  const stage = app.stage;
  stage.interactive = false;

  const labelsContainer = new Container<Text>({ zIndex: 3, isRenderGroup: true });
  const nodesContainer = new Container<Graphics>({ zIndex: 2, isRenderGroup: true });
  const linkContainer = new Container<Graphics>({ zIndex: 1, isRenderGroup: true });
  stage.addChild(nodesContainer, labelsContainer, linkContainer);

  // Create nodes
  for (const n of processedGraphData.nodes) {
    const nodeId = n.id;

    const label = new Text({
      interactive: false,
      eventMode: 'none',
      text: n.name,
      alpha: 0,
      anchor: { x: 0.5, y: 1.2 },
      style: {
        fontSize: fontSize * 15,
        fill: computedStyleMap['--dark'],
        fontFamily: computedStyleMap['--bodyFont'],
      },
      resolution: window.devicePixelRatio * 4,
    });
    label.scale.set(1 / scale);

    let oldLabelOpacity = 0;
    const isTag = isTagNode(n);
    const gfx = new Graphics({
      interactive: true,
      label: nodeId,
      eventMode: 'static',
      hitArea: new Circle(0, 0, nodeRadius(n)),
      cursor: 'pointer',
    })
      .circle(0, 0, nodeRadius(n))
      .fill({ color: isTag ? computedStyleMap['--light'] : color(n) })
      .on('pointerover', (e: any) => {
        updateHoverInfo(e.target.label);
        oldLabelOpacity = label.alpha;
        if (!dragging) {
          renderPixiFromD3();
        }
      })
      .on('pointerleave', () => {
        updateHoverInfo(null);
        label.alpha = oldLabelOpacity;
        if (!dragging) {
          renderPixiFromD3();
        }
      });

    if (isTag) {
      gfx.stroke({ width: 2, color: computedStyleMap['--tertiary'] });
    }

    nodesContainer.addChild(gfx);
    labelsContainer.addChild(label);

    const nodeRenderDatum: NodeRenderData = {
      simulationData: n,
      gfx,
      label,
      color: color(n),
      alpha: 1,
      active: false,
    };

    nodeRenderData.push(nodeRenderDatum);
  }

  // Create links
  for (const l of processedGraphData.links) {
    const gfx = new Graphics({ interactive: false, eventMode: 'none' });
    linkContainer.addChild(gfx);

    const linkRenderDatum: LinkRenderData = {
      simulationData: l,
      gfx,
      color: computedStyleMap['--lightgray'],
      alpha: 1,
      active: false,
    };

    linkRenderData.push(linkRenderDatum);
  }

  let currentTransform = zoomIdentity;

  // Handle drag
  if (enableDrag) {
    select<HTMLCanvasElement, NodeData | undefined>(app.canvas).call(
      drag<HTMLCanvasElement, NodeData | undefined>()
        .container(() => app.canvas)
        .subject(() => processedGraphData.nodes.find((n) => n.id === hoveredNodeId))
        .on('start', function dragstarted(event: any) {
          if (!event.active) simulation.alphaTarget(1).restart();
          event.subject.fx = event.subject.x;
          event.subject.fy = event.subject.y;
          event.subject.__initialDragPos = {
            x: event.subject.x,
            y: event.subject.y,
            fx: event.subject.fx,
            fy: event.subject.fy,
          };
          dragStartTime = Date.now();
          dragging = true;
        })
        .on('drag', function dragged(event: any) {
          const initPos = event.subject.__initialDragPos;
          event.subject.fx = initPos.x + (event.x - initPos.x) / currentTransform.k;
          event.subject.fy = initPos.y + (event.y - initPos.y) / currentTransform.k;
        })
        .on('end', function dragended(event: any) {
          if (!event.active) simulation.alphaTarget(0);
          event.subject.fx = null;
          event.subject.fy = null;
          dragging = false;

          if (Date.now() - dragStartTime < 500) {
            const node = processedGraphData.nodes.find(
              (n) => n.id === event.subject.id
            ) as NodeData;
            // Handle node click - you can customize this
            addToVisited(node.id);
            console.log('Node clicked:', node.id);
          }
        })
    );
  } else {
    for (const node of nodeRenderData) {
      node.gfx.on('click', () => {
        addToVisited(node.simulationData.id);
        console.log('Node clicked:', node.simulationData.id);
      });
    }
  }

  // Handle zoom
  if (enableZoom) {
    select<HTMLCanvasElement, NodeData>(app.canvas).call(
      zoom<HTMLCanvasElement, NodeData>()
        .extent([
          [0, 0],
          [width, height],
        ])
        .scaleExtent([0.25, 4])
        .on('zoom', ({ transform }: any) => {
          currentTransform = transform;
          stage.scale.set(transform.k, transform.k);
          stage.position.set(transform.x, transform.y);

          const scaleValue = transform.k * opacityScale;
          const scaleOpacity = Math.max((scaleValue - 1) / 3.75, 0);
          const activeNodes = nodeRenderData.filter((n) => n.active).flatMap((n) => n.label);

          for (const label of labelsContainer.children) {
            if (!activeNodes.includes(label)) {
              label.alpha = scaleOpacity;
            }
          }
        })
    );
  }

  // Animation loop
  let stopAnimation = false;
  function animate(time: number) {
    if (stopAnimation) return;

    for (const n of nodeRenderData) {
      const { x, y } = n.simulationData;
      if (!x || !y) continue;
      n.gfx.position.set(x + width / 2, y + height / 2);
      if (n.label) {
        n.label.position.set(x + width / 2, y + height / 2);
      }
    }

    for (const l of linkRenderData) {
      const linkData = l.simulationData;
      l.gfx.clear();
      l.gfx.moveTo(linkData.source.x! + width / 2, linkData.source.y! + height / 2);
      l.gfx
        .lineTo(linkData.target.x! + width / 2, linkData.target.y! + height / 2)
        .stroke({ alpha: l.alpha, width: 1, color: l.color });
    }

    tweens.forEach((t) => t.update(time));
    app.renderer.render(stage);
    requestAnimationFrame(animate);
  }

  requestAnimationFrame(animate);

  // Return cleanup function
  return () => {
    stopAnimation = true;
    app.destroy();
  };
}
