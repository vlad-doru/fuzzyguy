$.fn.serializeObject = function() {
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

var gauge_options = {
    width: 250,
    height: 250,
    min: 0,
    max: 100,
    yellowFrom: 70,
    yellowTo: 90,
    redFrom: 90,
    redTo: 100,
    minorTicks: 5,
    animation: {
        duration: 1000,
        easing: 'inAndOut',
    }
};

var correlation_options = {
    title: 'Accuracy and parameters correlation',
    animation: {
        duration: 1000,
        easing: 'out',
    },
    hAxis: {
        title: 'Levenshtein distance',
        minValue: 0,
        maxValue: 6
    },
    vAxis: {
        title: 'Accuracy',
        minValue: 0,
        maxValue: 100
    },
};

var performance_options = {
    title: 'Performance',
    curveType: 'function',
    legend: {
        position: 'bottom'
    },
    animation: {
        duration: 1000,
        easing: 'out',
    },
    hAxis: {
        title: 'Test size',
        minValue: 0,
        // maxValue: 600000
    },
    vAxis: {
        title: 'Performance (ms)',
        minValue: 0,
    },
    trendlines: {}
};

(function() {
    google.load("visualization", "1", {
        packages: ["corechart", "gauge"]
    });
    google.setOnLoadCallback(loadData);

    function loadData() {
        correlationchart = new google.visualization.BubbleChart(document.getElementById('correlation'));
        performancechart = new google.visualization.ScatterChart(document.getElementById('performance'));

        accuracy = new google.visualization.Gauge(document.getElementById('accuracygauge'))
        var data = google.visualization.arrayToDataTable([
            ['Label', 'Value'],
            ['Accuracy', 0],
        ])
        accuracy.draw(data, gauge_options)
    }
})()

function newPerfDataSet(rows) {
    var perfdata = new google.visualization.DataTable();
    perfdata.addColumn('number', 'Size')
    perfdata.addColumn('number', 'Distance of 1')
    perfdata.addColumn('number', 'Distance of 2')
    perfdata.addColumn('number', 'Distance of 3')
    perfdata.addRows(rows)
    return perfdata
}

$(document).ready(function() {
    var testform = $("#test")
    var spin = $(".spinner")
    var warning = $("#warning")
    var set = $("#set")
    var CorrelationData = [
        ['ID', 'Levenshtein Distance', 'Accuracy', 'Test size', 'Numer of results'],
    ]
    var PerformanceData = []
    var TestNr = 0
    testform.on('submit', function(e) {
        spin.toggleClass('paused')
        warning.slideDown(300)
        spin.animate({
            height: '30px',
            "margin-top": '5px'
        }, 300)
        e.preventDefault();
        var test_options = testform.serializeObject()

        $.ajax({
            "url": "/fuzzy/test",
            "type": 'GET',
            dataType: 'json',
            data: test_options,
            success: function(data) {
                for (var prop in data) {
                    $("#" + prop).html(data[prop])
                }

                var datachart = google.visualization.arrayToDataTable([
                    ['Label', 'Value'],
                    ['Accuracy', data['accuracy']],
                ])
                accuracy.draw(datachart, gauge_options)
                $("#accuracygauge").fadeIn()
                TestNr++
                console.log(data)

                CorrelationData.push(['Testul #' + TestNr, Number(data['distance']), Number(data['accuracy']), String(data['keys']), Number(data['results'])])
                var perf = []
                for (var i = 3; i >= 0; i--) {
                    perf[i] = null
                };
                perf[0] = Number(data['keys'])
                perf[data['distance']] = Number(data['time'])
                performance_options['trendlines'][data['distance'] - 1] = {}
                PerformanceData.push(perf)
                console.log(PerformanceData)

                var datachart = google.visualization.arrayToDataTable(CorrelationData)
                correlationchart.draw(datachart, correlation_options)
                perfdata = newPerfDataSet(PerformanceData)
                performancechart.draw(perfdata, performance_options)
                $("#correlation").fadeIn()
                $("#performance").fadeIn()

                set.fadeIn()
            },
            error: function(jqXHR, textStatus, errorThrown) {
                console.log(jqXHR)
                console.log(errorThrown)
            }
        }).always(function() {
            spin.toggleClass('paused')
            spin.animate({
                height: '0px',
                "margin-top": '15px'
            }, 300)
            warning.slideUp(300)
        })
    })
})
