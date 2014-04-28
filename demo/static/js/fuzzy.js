$(document).ready(function(){
		var serviceURL = "fuzzy?store=demostore&distance=2&results=5&key="
		var exactURL = "fuzzy?store=demostore&distance=0&key="
		var fuzzyURL = "fuzzy"
		
		var input = $("#english")
		var definition = $("#definition")
		var dictionary = $("#dictionary")
		var addword = $("#addword")

		$.ajax({
			"url":fuzzyURL,
			"type": 'POST',
			"data": {"store": "demostore"}
		});

		input.autocomplete({
			source: function( request, response ){
				var url =  serviceURL + request["term"]
				url = encodeURI(url)
				$.getJSON(url, response);
			}
    	});
    	input.on('change paste keyup', function(){
    		var url = exactURL + input.val()
    		url = encodeURI(url)
    		$.ajax({
   				"url":url,
    			"type": 'GET',
    			success: function(data){
    				definition.text(data)
    			},
    			error: function(data){
    				definition.text("")
    			}
    		});
    	})

    	dictionary.on('submit', function(e){
    		e.preventDefault();
    	})

    	addword.on('submit', function(e){
    		e.preventDefault();
    		var key = $("#key").val()
    		var value = $("#value").val()
    		var dict = {
    			key: key,
    			value: value,
    			store: "demostore"
    		}
    		$.ajax({
   				"url":fuzzyURL,
    			"type": 'PUT',
    			"data": dict,
    			success: function(data){
    				addword.find("input[type=text], textarea").val("")
    				$("#success").slideDown().delay(1500).slideUp()
    			},
    			error: function(data){
    				console.log(data)
    			}
    		});
    	});

    	var english = $("#loadenglish")
    	// Ladda.bind('#loadenglish')

    	english.on('click', function(e){
    		e.preventDefault()
    		var l = Ladda.create(this)
    		l.start()
    		$.ajax({
   				"url": "demo/loadenglish",
    			"type": 'GET',
    			success: function(data){
    				console.log(data)
    				$("#engsuccess").slideDown().delay(1500).slideUp()
    			},
    			error: function(data){
    				console.log(data)
    			}
    		}).always(function(){
    			l.stop()
    		})
    	})

    	var histhead = $("#histhead")
    	var histbody = $("#histbody")
    	for (var i = 0; i < 32; i++) {
    		histhead.append("<th>"+i+"</th>")
    		histbody.append("<td>0</td>")
    	};

    	var values = $("#histbody td")
    	var word = $("#word strong")

    	function sleep(millis, callback, j) {
		    setTimeout(function()
		            { callback(j); }
		    , millis);
		}
		var hash=$("#hash")
		var charduration = 2000

		var computeHist = function(id) {
			w = $("#word" + id).val()
			for (var i =0 ; i<w.length; i++) {
				sleep(i*charduration, function(j){
					var prefix = w.substring(0, j)
					var suffix = w.substring(j+1, w.length)
					word.html(prefix + "<span style='border-bottom:3px solid white; text-algin:center'>"+w[j]+"</span>" + suffix)
					var code = w[j].charCodeAt()
					var bucket = code%32
					hash.html( code + "(" + w[j] + ") % 32 = " + (bucket))
					var cell = $(values[bucket])
					var add_to = parseInt(cell.text())
					cell.css('background-color', '#5bc0de')
					cell.text((add_to + 1)%2)
					setTimeout(function(){
						cell.css('background-color', 'transparent')
					},1500)
				}, i)
    		} 
    		setTimeout(function(){
    			word.html("&zwnj;")	
    			var hist = ""
    			for (var i =0; i < values.length; i++){
    				cell = $(values[i])
    				hist += cell.text()
    				cell.text("0")
    			}	
    			$("#hist" + id).text(hist)
    			hash.html("&zwnj;")
    		}, w.length*charduration) 
		}

    	$("#histogram").on('submit', function(e){
    		$("#hist1").html("&zwnj;")
    		$("#hist2").html("&zwnj;")
    		$("#x1").html("&zwnj;")
    		$("#x2").html("&zwnj;")
    		$("#difference").html("");
    		e.preventDefault();
    		computeHist("1") 
    		setTimeout(function(){
    			computeHist("2") 
    		}, ($("#word1").val().length+1) * charduration);
    		setTimeout(function(){
    			$("#x1").html($("#hist1").text())
    			$("#x1").fadeTo(600, 0.8)
    			setTimeout(function(){
    				$("#x2").html($("#hist2").text())
    				$("#x2").fadeTo(600, 0.8)
    			},1000)
    			setTimeout(function(){
    				var dif = 0
    				var h1 = $("#hist1").text()    				
    				var h2 = $("#hist2").text()
    				for (var i = 0 ; i < 32; ++i){
    					if (h1[i] != h2[i]) {
    						dif++
    					}
    				}
    				var lower = Math.abs($("#word1").val().length - $("#word2").val().length)
    				$("#difference").html("Biti diferiti " + dif + "<br/> Limita inferioara este (" + dif + "(biti diferiti) + " + lower + "(diferenta lungimi) ) / 2 = " + (lower + dif)/2).slideDown()
    			})
    		}, charduration*($("#word1").val().length+$("#word2").val().length+2))
    	});
	})