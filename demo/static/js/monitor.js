$.fn.serializeObject = function()
{
    var o = {};
    var a = this.serializeArray();
    $.each(a, function() {
        if (o[this.name] !== undefined) {
            if (!o[this.name].push) {
                o[this.name] = [o[this.name]];
            }
            o[this.name].push(this.value || '');
        } else {
            o[this.name] = this.value || '';
        }
    });
    return o;
};

(function(){
	google.load("visualization", "1", {packages:["corechart"]});
	google.setOnLoadCallback(loadData);
	function loadData(){
		var data = google.visualization.arrayToDataTable([
			['Testare', 'Dimensiune', 'Timp'],
			// ['1', 100000, 985]
		])
		var options = {
			title: 'Test'
		}

		var chart = new google.visualization.LineChart(document.getElementById('performance'));
        // chart.draw(data, options);
	}
})()

$(document).ready(function(){
	var testform = $("#test")
	var spin = $(".spinner")
	var warning = $("#warning")
	var set = $("#set")
	testform.on('submit', function(e){
		spin.toggleClass('paused')
		warning.slideDown(300)
		spin.animate({
			height: '30px',
			"margin-top": '5px'
		}, 300)
		e.preventDefault();
		var test_options = testform.serializeObject()
		console.log(test_options)
		$.ajax({
			"url": "/fuzzy/test",
			"type": 'GET',
			dataType: 'json',
			data: test_options,
			success: function(data){
				for (var prop in data) {
					$("#"+prop).html(data[prop])
				}
				set.show()
			},
			error: function(jqXHR, textStatus, errorThrown){
				console.log(jqXHR)
				console.log(errorThrown)
			}
		}).always(function(){
			spin.toggleClass('paused')
			spin.animate({
			height: '0px',
			"margin-top": '15px'
		}, 300)
		warning.slideUp(300)
		})
	})
})