# DagView

Takes SVG output from GraphViz, and returns HTML with some JS and CSS
injected into that makes it interactive.

When you open the resulting HTML in a browser and hover over a node:

- The node gets highlighted red
- The edges coming out of it become red
- The edges coming into it become green
- The nodes on the end of out edges are circled in red
- The nodes on the end of in edges are circled in green

It looks like this:

![dag view](./example/example.png)
[Live Demo](https://amazing-feynman-ce1493.netlify.com/example.svg)

## Usage

```bash
cat myfile.dot \
  | dot -Tsvg \
  | dagview \
  > out.html
```
 
...and open up `out.svg` in your browser.
