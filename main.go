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
	if err := insertScriptAndStyle(tree); err != nil {
		log.Fatal("failed inserting script and style: ", err)
	}
	_, err = tree.WriteTo(os.Stdout)
	if err != nil {
		log.Fatal("error writing output: ", err)
	}
}

func insertScriptAndStyle(tree *etree.Document) error {
	styleEl := etree.NewElement("style")
	styleEl.AddChild(etree.NewCData(Style))

	scriptEl := etree.NewElement("script")
	scriptEl.AddChild(etree.NewCData(Script))

	for _, c := range tree.Child {
		switch el := c.(type) {
		case *etree.Element:
			if el.Tag == "svg" {
				el.AddChild(styleEl)
				el.AddChild(scriptEl)
				return nil
			}
		}
	}
	return fmt.Errorf("didn't find an <svg> element at the root")
}

const Style = `
.node path {
    fill: #afeeee;
    stroke: #afeeee;
}
.node.selected path {
    fill: red;
    stroke: red;
}
.node.selected-in path {
    stroke: green;
}
.node.selected-out path {
    stroke: red;
}

.edge path {
    stroke: black;
}
.edge.selected-in path {
    stroke: green;
}
.edge.selected-out path {
    stroke: red;
}
`

const Script = `
function addListenersToNodes() {
		const nodes = document.querySelectorAll(".node");
		nodes.forEach(node => {
				node.addEventListener("mouseenter", evt => {
						const name = node.querySelector("title").innerHTML;
						updateSelection(name, node);
				});
				// node.addEventListener("mouseleave", evt => {
				//     reset();
				// });
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
				const title = edge.querySelector("title").innerHTML;
				const [from, to] = title.split("-&gt;");
				addEdge(outEdges, from, to, edge);
				addEdge(inEdges, to, from, edge);
		});
		const nodeNodes = document.querySelectorAll(".node");
		const nodes = {}; // node name => node DOM nodee
		nodeNodes.forEach(node => {
				const title = node.querySelector("title").innerHTML;
				nodes[title] = node;
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
}

function updateSelection(name, node) {
		reset();

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

		console.log(
				"selected:", name,
				"out edges:", outFromThis.map(n => n.to),
				"in edges:", inToThis.map(n => n.to),
		);
}

const { nodes, outEdges, inEdges } = edgesAndNodesByName();
addListenersToNodes();
`
