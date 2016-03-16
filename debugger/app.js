$(document).ready(function() {
	// Websocket
	var socket;
	$('#disconnect').click(function(){
		socket.close();
	});

	$('#connect').click(function(){
		try{
			socket = new WebSocket("ws://127.0.0.1:8888/socket");
			message('<div class="info event">Socket Status: </div><div class="data">', socket.readyState);

			socket.onopen = function(){
				message('<div class="info event">Socket Status (open): </div><div class="data">', socket.readyState);
			}

			socket.onmessage = function(msg){
				message('<div class="info ws-message">Received: </div><div class="data">', msg.data);
			}

			socket.onclose = function(){
				message('<div class="info event">Socket Status (closed): </div><div class="data">', socket.readyState);
			}

		} catch(exception){
			message('<p>Error', exception);
		}
    });

	function message(msg, obj){
		$('#chatLog').append('<div class="data-wrapper">'+msg+JSON.stringify(JSON.parse(obj), null, 2)+'</div></div>');
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
	var height = 350;

	//The number of columns and rows of the heatmap
	var MapColumns = 30,
		MapRows = 20;
		
	//The maximum radius the hexagons can have to still fit the screen
	var hexRadius = d3.min([width/((MapColumns + 0.5) * Math.sqrt(3)),
				height/((MapRows + 1/3) * 1.5)]);

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
	    .attr("class", "hexagon")
	    .attr("d", function (d) {
			return "M" + d.x + "," + d.y + hexbin.hexagon();
		})
	    .attr("stroke", function (d,i) {
			return "#fff";
		})
	    .attr("stroke-width", "1px")
	    .style("fill", function (d,i) {
			return color[i];
		})
		.on("mouseover", mover)
		.on("mouseout", mout)
		;
});