$(document).ready(function() {
	var socket;

	// Api
	$('#start').click(function() {
		$.post('/race/start',
			function(data) {
	    		message('<div class="info message">API start: </div><div class="data">', data);
	    	},
	    	function(err) {
	    		message('<div class="info warning">API start: </div><div class="data err">', err);
	    	}
    	);
  	});

	$('#stop').click(function() {
    	$.post('/race/stop',
    		function(data) {
	    		message('<div class="info message">API stop: </div><div class="data">', data);
	    	},
	    	function(err) {
	    		message('<div class="info warning">API stop: </div><div class="data err">', err);
	    	}
	    );
  	});

  	$('#get_race').click(function() {
    	$.get('/race/'+$('#race_id').val(),
    		function(data) {
	    		message('<div class="info message">API race: </div><div class="data">', data);
	    	},
	    	function(err) {
	    		message('<div class="info warning">API race: </div><div class="data err">', err);
	    	}
	    );
  	});

	$('#get_races').click(function() {
    	$.get('/races',
    		function(data) {
	    		message('<div class="info message">API races: </div><div class="data">', data);
	    	},
	    	function(err) {
	    		message('<div class="info warning">API races: </div><div class="data err">', err);
	    	}
	    );
  	});

	$('#get_transponders').click(function() {
    	$.get('/transponders',
    		function(data) {
	    		message('<div class="info message">API transponders: </div><div class="data">', data);
	    	},
	    	function(err) {
	    		message('<div class="info warning">API transponders: </div><div class="data err">', err);
	    	}
	    );
  	});

  	$('#put_transponders').click(function() {
  		var obj = [{id: 2596996162, car_num: 4}]
    	$.ajax({
			url: '/transponders',
			type: 'PUT',
			contentType: 'application/json',
			data: JSON.stringify(obj),
			success: function(data) {
	    		message('<div class="info message">API transponders: </div><div class="data">', data);
	    	},
	    	error: function(err) {
	    		message('<div class="info warning">API transponders: </div><div class="data err">', err);
	    	}
	    });
  	});

  	$('#put_drivers').click(function() {
  		var obj = [{
  					race_id: parseInt($('#d_race_id').val()),
  					transponder_id: parseInt($('#d_t_id').val()),
  					car_num: parseInt($('#d_car_num').val()),
  					driver_name: $('#d_name').val()
  				}]
    	$.ajax({
			url: '/drivers',
			type: 'PUT',
			contentType: 'application/json',
			data: JSON.stringify(obj),
			success: function(data) {
	    		message('<div class="info message">API drivers: </div><div class="data">', data);
	    	},
	    	error: function(err) {
	    		message('<div class="info warning">API drivers: </div><div class="data err">', err);
	    	}
	    });
  	});

  	$('#get_top_all').click(function() {
    	$.get('/top/all/10',
    		function(data) {
	    		message('<div class="info message">API race: </div><div class="data">', data);
	    	},
	    	function(err) {
	    		message('<div class="info warning">API race: </div><div class="data err">', err);
	    	}
	    );
  	});

  	$('#get_top_year').click(function() {
    	$.get('/top/year/10',
    		function(data) {
	    		message('<div class="info message">API race: </div><div class="data">', data);
	    	},
	    	function(err) {
	    		message('<div class="info warning">API race: </div><div class="data err">', err);
	    	}
	    );
  	});

  	$('#get_top_month').click(function() {
    	$.get('/top/month/10',
    		function(data) {
	    		message('<div class="info message">API race: </div><div class="data">', data);
	    	},
	    	function(err) {
	    		message('<div class="info warning">API race: </div><div class="data err">', err);
	    	}
	    );
  	});

	// Websocket
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
});