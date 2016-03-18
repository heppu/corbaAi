$(document).ready(function() {
	var socket;
	var grid = {};

	connect();

	function connect() {
		try{
			socket = new WebSocket("ws://localhost:8888/socket");

			socket.onopen = function(){
				console.log("Socket opened");
			}

			socket.onmessage = function(msg){
				var res = JSON.parse(msg.data);
				parseResponse(res);
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
		var bots = res.bots || [];
		var map = res.map || [];

		for (var i=0; i<map.length; i++) {
			// If there's change to previously stored value
			if (grid[map[i].x + "_" + map[i].y] || grid[map[i].x + "_" + map[i].y] !== map[i].probed) {
				// If area has not been discovered
				if (!map[i].probed) {
					$("#" + map[i].x + "_" + map[i].y).css("fill", "#002672");
				} else { // Area discovered
					// If enemy bots have been discovered in the area
					if (map[i].bots.length) {
						$("#" + map[i].x + "_" + map[i].y).css("fill", "#B80000");	
					} else { // Empty area
						$("#" + map[i].x + "_" + map[i].y).css("fill", "#95B2D2");	
					}
				}
			} else {
				$("#" + map[i].x + "_" + map[i].y).css("fill", "#002672");
			}
			grid[map[i].x + "_" + map[i].y] = map[i].probed;
		}



		for (var i=0; i<bots.length; i++) {
			if (bots[i].alive) {
				$("#" + bots[i].pos.x + "_" + bots[i].pos.y).css("fill", "#00CC2E")
			} else {
				$("#" + bots[i].pos.x + "_" + bots[i].pos.y).css("fill", "#A81013")
			}
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
	var width = 700;
	var height = 700;

	//The number of columns and rows of the heatmap
	var MapColumns = 2*14+1,
		MapRows = 2*14+1;

	// Size of hexagon
	//The maximum radius the hexagons can have to still fit the screen
	//var hexRadius = d3.min([width/((MapColumns + 0.5) * Math.sqrt(3)),
	//			height/((MapRows + 1/3) * 1.5)]);
	var hexRadius = 17;

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