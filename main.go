package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/beevik/etree"
)

func main() {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("error reading stdin: ", err)
	}
	tree := etree.NewDocument()
	err = tree.ReadFromBytes(bytes)
	if err != nil {
		log.Fatal("error parsing XML: ", err)
	}
	doc := etree.NewDocument()
	g := getSvgChild(tree)
	doc.Element = *g
	str, err := doc.WriteToString()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(template, str, Script, Style)
}

const template = `
<!DOCTYPE html>
<html>
	<body style="margin: 0">
		<svg id="viz" style="width: 100%%; height: 100vh;">
			%s
		</svg>
		<script id="panel-template" type="text/x-handlebars-template">
			{{#if nodeInfo}}
				<div id="selected-area">
					<h2>Selected</h2>
					<span class="node-name">{{nodeInfo.selected}}</span>
				</div>
				<div id="in-edges-list">
					<h2>In Edges</h2>
					<ul>
						{{#each nodeInfo.inEdges}}
							<li
								class="node-name"
								onClick="updateSelection('{{to}}')"
								onMouseOver="showHovered('{{to}}')"
								onMouseOut="unShowHovered('{{to}}')"
							>
								{{to}}
							</li>
						{{/each}}
					</ul>
				</div>
				<div id="out-edges-list">
					<h2>Out Edges</h2>
					<ul>
						{{#each nodeInfo.outEdges}}
							<li
								class="node-name"
								onClick="updateSelection('{{to}}')"
								onMouseOver="showHovered('{{to}}')"
								onMouseOut="unShowHovered('{{to}}')"
							>
								{{to}}
							</li>
						{{/each}}
					</ul>
				</div>
			{{/if}}
			{{#if graphInfo}}
				<h2>Tops</h2>
				<ul>
					{{#each graphInfo.tops}}
						<li
							class="node-name"
							onClick="updateSelection('{{.}}')"
							onMouseOver="showHovered('{{.}}')"
							onMouseOut="unShowHovered('{{.}}')"
						>
							{{.}}
						</li>
					{{/each}}
				</ul>
				<h2>Bottoms</h2>
				<ul>
					{{#each graphInfo.bottoms}}
						<li
							class="node-name"
							onClick="updateSelection('{{.}}')"
							onMouseOver="showHovered('{{.}}')"
							onMouseOut="unShowHovered('{{.}}')"
						>
							{{.}}
						</li>
					{{/each}}
				</ul>
			{{/if}}
		</script>
		<div id="panel">
		</div>
	</body>
	<script src="https://cdn.jsdelivr.net/npm/svg-pan-zoom@3.5.0/dist/svg-pan-zoom.min.js"></script>
	<script src="https://cdn.jsdelivr.net/npm/handlebars@4.1.1/dist/handlebars.min.js"></script>
	<script>
		%s
	</script>
	<style>
		%s
	</style>
</html>
`

func getSvgChild(tree *etree.Document) *etree.Element {
	path := etree.MustCompilePath("/svg/g")
	return tree.FindElementPath(path)
}

const Style = `
#panel {
		position: absolute;
		height: 100vh;
		right: 0;
		top: 0;
		width: 300px;
		background: rgb(220, 220, 220, 0.9);
		padding-left: 10px;
		padding-right: 10px;
		overflow: scroll;
}
.node-name {
		font-family: monospace;
		cursor: pointer;
}
.node-name:hover {
		color: orange;
}
h2 {
		font-family: sans-serif;
}
#selected-area {
		color: red;
}
#in-edges-list {
		color: green;
}
#out-edges-list {
		color: blue;
}

.node polygon {
    fill: #afeeee;
    stroke: #afeeee;
}
.node.hovered polygon,
.node.selected-in.hovered polygon,
.node.selected-out.hovered polygon {
		fill: orange;
		stroke: orange;
}
.node.selected polygon {
    fill: red;
    stroke: red;
}
.node.selected text {
  	fill: white;
}
.node.selected-in polygon {
    stroke: green;
		fill: green;
}
.node.selected-in text {
  	fill: white;
}
.node.selected-out polygon {
    stroke: blue;
		fill: blue;
}
.node.selected-out text {
  fill: white;
}

.edge path {
    stroke: black;
}
.edge.selected-in path {
    stroke: green;
}
.edge.selected-out path {
    stroke: blue;
}
`

const Script = `
const panelTemplate = Handlebars.compile(document.getElementById("panel-template").innerHTML);

function addListenersToNodes() {
		const nodes = document.querySelectorAll(".node");
		nodes.forEach(node => {
				node.addEventListener("click", evt => {
						evt.preventDefault();
						const name = node.getAttribute("id");
						updateSelection(name);
				});
		});
}

function addEdge(obj, from, to, node) {
		let edges = obj[from];
		if (!edges) {
				edges = [];
				obj[from] = edges;
		}
		edges.push({
				node,
				to,
				from,
		});
}

function edgesAndNodesByName() {
		const outEdges = {}; // node name (string) => [edge DOM node]
		const inEdges = {}; // node name (string) => [edge DOM node]
		const edgeNodes = document.querySelectorAll(".edge");
		edgeNodes.forEach(edge => {
				const title = edge.getAttribute("id");
				const [from, to] = title.split("->");
				addEdge(outEdges, from, to, edge);
				addEdge(inEdges, to, from, edge);
		});
		const nodeNodes = document.querySelectorAll(".node");
		const nodes = {}; // node name => node DOM nodee
		nodeNodes.forEach(node => {
				const name = node.getAttribute("id");
				nodes[name] = node;
		});
		return { nodes, outEdges, inEdges };
}

let selectedNode = null;
let depNodes = [];
let selectedEdges = [];

function reset() {
		if (selectedNode) {
				selectedNode.classList.remove("selected");
		}
		selectedEdges.forEach(edge => {
				edge.classList.remove("selected-in");
				edge.classList.remove("selected-out");
		});
		selectedEdges = [];
		depNodes.forEach(node => {
				node.classList.remove("selected-in");
				node.classList.remove("selected-out");
		});
		depNodes = [];
		updateSidebar({});
}

function updateSelection(name) {
		reset();

		history.pushState({ name }, name, "#" + name);

		const node = nodes[name];
		if (node === selectedNode) {
				selectedNode = null;
				return;
		}

		node.classList.add("selected");
		selectedNode = node;

		// highlight edges
		const outFromThis = outEdges[name] || [];
		outFromThis.forEach(outEdge => {
				outEdge.node.classList.add("selected-out");
				selectedEdges.push(outEdge.node);
				const node = nodes[outEdge.to];
				node.classList.add("selected-out");
				depNodes.push(node);
		});
		const inToThis = inEdges[name] || [];
		inToThis.forEach(inEdge => {
				inEdge.node.classList.add("selected-in");
				selectedEdges.push(inEdge.node);
				const node = nodes[inEdge.to];
				node.classList.add("selected-in");
				depNodes.push(node);
		});
		
		updateSidebar({
				nodeInfo: {
						selected: name,
						inEdges: inToThis,
						outEdges: outFromThis,
				},
		});
}

function updateSidebar(info) {
		if (!info.nodeInfo) {
				info.graphInfo = {
						tops: tops,
						bottoms: bottoms,
				};
		}
		document.getElementById("panel").innerHTML = panelTemplate(info);
}

function getTopsAndBottoms(nodes, outEdges, inEdges) {
		const tops = [];
		const bottoms = [];
		Object.keys(nodes).forEach((name) => {
				const outFromThis = outEdges[name] || [];
				const inToThis = inEdges[name] || [];
				if (outFromThis.length === 0) {
						bottoms.push(name);
				}
				if (inToThis.length === 0) {
						tops.push(name);
				}
		});
		return { tops, bottoms };
}

function showHovered(name) {
		nodes[name].classList.add("hovered");
}

function unShowHovered(name) {
		nodes[name].classList.remove("hovered");
}

const { nodes, outEdges, inEdges } = edgesAndNodesByName();
const { tops, bottoms } = getTopsAndBottoms(nodes, outEdges, inEdges);
updateSidebar({});
addListenersToNodes();
window.addEventListener("popstate", (evt) => {
		console.log("popstate", evt.state);
		if (evt.state.name) {
				updateSelection(evt.state.name);
		}
});
svgPanZoom("#viz", {
		maxZoom: 100
});
`
