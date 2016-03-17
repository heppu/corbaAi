$(document).ready(function() {
	var socket;
	var grid = {};

	connect();

	function connect() {
		try{
			socket = new WebSocket("ws://localhost:9999/socket");

			socket.onopen = function(){
				console.log("Socket opened");
			}

			socket.onmessage = function(msg){
				var res = JSON.parse(msg.data);
				if ('map' in res) {
					parseResponse(res.map);
				}
			}

			socket.onclose = function() {
				console.log("Socket closed");
			}

		} catch(exception) {
			console.log("Fug", exception);
		}
	}

	function disconnect() {
		socket.close();
	}

	function parseResponse(res) {
		for (var i=0; i<res.length; i++) {
			if (grid[res[i].x + "_" + res[i].y] || grid[res[i].x + "_" + res[i].y] !== res[i].empty) {
				if (!res[i].empty) {
					$("#" + res[i].x + "_" + res[i].y).css("fill", "#002672");
				} else {
					$("#" + res[i].x + "_" + res[i].y).css("fill", "#95B2D2");
				}
			}
			grid[res[i].x + "_" + res[i].y] = res[i].empty;
		}
	}


	// Hexagon
	//svg sizes and margins
	var margin = {
	    top: 30,
	    right: 20,
	    bottom: 20,
	    left: 50
	};

	//The next lines should be run, but this seems to go wrong on the first load in bl.ocks.org
	//var width = $(window).width() - margin.left - margin.right - 40;
	//var height = $(window).height() - margin.top - margin.bottom - 80;
	//So I set it fixed to
	var width = 850;
	var height = 850;

	//The number of columns and rows of the heatmap
	var MapColumns = 2*14+1,
		MapRows = 2*14+1;

	// Size of hexagon
	//The maximum radius the hexagons can have to still fit the screen
	//var hexRadius = d3.min([width/((MapColumns + 0.5) * Math.sqrt(3)),
	//			height/((MapRows + 1/3) * 1.5)]);
	var hexRadius = 20;

	//Set the new height and width of the SVG based on the max possible
	width = MapColumns*hexRadius*Math.sqrt(3);
	heigth = MapRows*1.5*hexRadius+0.5*hexRadius;

	//Set the hexagon radius
	var hexbin = d3.hexbin()
	    	       .radius(hexRadius);

	//Calculate the center positions of each hexagon
	var points = [];
	for (var i = 0; i < MapRows; i++) {
	    for (var j = 0; j < MapColumns; j++) {
	        points.push([hexRadius * j * 1.75, hexRadius * i * 1.5]);
	    }//for j
	}//for i

	//Create SVG element
	var svg = d3.select("#chart").append("svg")
	    .attr("width", width + margin.left + margin.right)
	    .attr("height", height + margin.top + margin.bottom)
	    .append("g")
	    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

	//Start drawing the hexagons
	svg.append("g")
	    .selectAll(".hexagon")
	    .data(hexbin(points))
	    .enter().append("path")
	    .attr("id", function(d) { return (Math.ceil(-14/2-(d.j/2))+d.i) + "_" + (-14+d.j); })
	    .attr("alt", function(d) { return (Math.ceil(-14/2-(d.j/2))+d.i) + "_" + (-14+d.j); })
	    .attr("class", "hexagon")
	    .attr("d", function (d) {
			return "M" + d.x + "," + d.y + hexbin.hexagon();
		})
	    .attr("stroke", function (d,i) {
			return "#000";
		})
	    .attr("stroke-width", "1px")
	    .style("fill", "F6F8FB");
});