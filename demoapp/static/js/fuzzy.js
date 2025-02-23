$(document).ready(function() {
    var serviceURL = "/fuzzy/query?key="
    var exactURL = "fuzzy/exact?key="

    // Elements which need to be taken care of
    var input = $("#english")
    var clear = $("#clear")
    var definition = $("#definition")
    var dictionary = $("#dictionary")
    var addword = $("#addword")
    var english = $("#loadenglish")

    input.autocomplete({
        source: function(request, response) {
            var url = serviceURL + request["term"]
            url = encodeURI(url)
            $.getJSON(url, response);
        }
    });
    input.on('change paste keyup', function() {
        var url = exactURL + input.val()
        url = encodeURI(url)
        $.ajax({
            "url": url,
            "type": 'GET',
            success: function(data) {
                definition.text(data)
            },
            error: function(data) {
                definition.text("")
            }
        });
    })

    dictionary.on('submit', function(e) {
        e.preventDefault();
    })

    addword.on('submit', function(e) {
        e.preventDefault();
        var key = $("#key").val()
        var value = $("#value").val()
        var dict = {
            key: key,
            value: value,
            store: "demostore"
        }
        $.ajax({
            "url": "/fuzzy/add",
            "type": 'PUT',
            "data": dict,
            success: function(data) {
                addword.find("input[type=text], textarea").val("")
                $("#success").slideDown().delay(1500).slideUp()
            },
            error: function(data) {
                console.log(data)
            }
        });
    });

    clear.on('click', function(e) {
        e.preventDefault()
        $.ajax({
            "url": "fuzzy/cleardemo",
            "type": 'GET',
            success: function(data) {
                console.log(data)
                $("#clearsuccess").slideDown().delay(1500).slideUp()
            },
            error: function(data) {
                console.log(data)
            }
        })
    })


    english.on('click', function(e) {
        e.preventDefault()
        var l = Ladda.create(this)
        l.start()
        $.ajax({
            "url": "fuzzy/loadenglish",
            "type": 'GET',
            success: function(data) {
                console.log(data)
                $("#engsuccess").slideDown().delay(1500).slideUp()
            },
            error: function(data) {
                console.log(data)
            }
        }).always(function() {
            l.stop()
        })
    })

    var histhead = $("#histhead")
    var histbody = $("#histbody")
    for (var i = 0; i < 32; i++) {
        histhead.append("<th>" + i + "</th>")
        histbody.append("<td>0</td>")
    };

    var values = $("#histbody td")
    var word = $("#word strong")

        function sleep(millis, callback, j) {
            setTimeout(function() {
                callback(j);
            }, millis);
        }
    var hash = $("#hash")
    var charduration = 2000

    var computeHist = function(id) {
        w = $("#word" + id).val()
        word.fadeIn()
        for (var i = 0; i < w.length; i++) {
            sleep(i * charduration, function(j) {
                var prefix = w.substring(0, j)
                var suffix = w.substring(j + 1, w.length)
                word.html(prefix + "<span style='border-bottom:3px solid white; text-algin:center'>" + w[j] + "</span>" + suffix)
                var code = w[j].charCodeAt()
                var bucket = code % 32
                hash.html(code + "(" + w[j] + ") % 32 = " + (bucket))
                var cell = $(values[bucket])
                var add_to = parseInt(cell.text())
                cell.animate({
                    'background-color': '#5bc0de'
                }, 700)
                cell.text((add_to + 1) % 2)
                setTimeout(function() {
                    cell.animate({
                        'background-color': 'transparent'
                    }, 700);
                }, 1500)
            }, i)
        }
        setTimeout(function() {
            word.html("&zwnj;")
            var hist = ""
            for (var i = 0; i < values.length; i++) {
                cell = $(values[i])
                hist += cell.text()
                cell.text("0")
            }
            $("#hist" + id).hide().text(hist).slideDown()
            hash.html("&zwnj;")
        }, w.length * charduration)
    }

    $("#histogram").on('submit', function(e) {
        $("#hist1").html("&zwnj;")
        $("#hist2").html("&zwnj;")
        $("#x1").html("&zwnj;")
        $("#x2").html("&zwnj;")
        $("#difference").html("");
        e.preventDefault();
        computeHist("1")
        setTimeout(function() {
            computeHist("2")
        }, ($("#word1").val().length + 1) * charduration);
        setTimeout(function() {
            $("#x1").html($("#hist1").text())
            $("#x1").slideDown()
            setTimeout(function() {
                $("#x2").html($("#hist2").text())
                $("#x2").slideDown()
                word.slideUp()
            }, 1000)
            setTimeout(function() {
                var dif = 0
                var h1 = $("#hist1").text()
                var h2 = $("#hist2").text()
                for (var i = 0; i < 32; ++i) {
                    if (h1[i] != h2[i]) {
                        dif++
                    }
                }
                var lower = Math.abs($("#word1").val().length - $("#word2").val().length)
                $("#difference").html("Biti diferiti " + dif + "<br/> Limita inferioara este (" + dif + "(biti diferiti) + " + lower + "(diferenta lungimi) ) / 2 = " + (lower + dif) / 2).slideDown()
            })
        }, charduration * ($("#word1").val().length + $("#word2").val().length + 2))
    });
})
