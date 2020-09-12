
(function(){
    let schema = "ws";
    if (location.protocol === "https:") schema = "wss";
    let ws_url = schema+'://'+location.hostname+(location.port ? ':'+location.port: '')+"/ws";
    let domain = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');

    let doms = document.getElementsByClassName("domain-name");
    for(let i = 0; i < doms.length; i++){
        doms[i].innerHTML = domain;
    }

    let socket = null;
    let msg = JSON.stringify({
        name: "subscribe",
        payload: "all",
    });
    let html = "";
    let list = [];
    let subscribed = false;

    function connect_socket() {
        socket = new WebSocket(ws_url);
        socket.onclose = function(event){
            connect_socket();
        };
        socket.onopen = function(event){
            socket.send(msg);
        };
        listen();
    }

    function fix_number(number) {
        if (number <= 9) return "0" + number;
        return number;
    }

    function get_date(date) {
        return fix_number(date.getHours()) + ":" + fix_number(date.getMinutes()) + ":" + fix_number(date.getSeconds())
    }

    function listen() {
        socket.onmessage = function (event) {
            if (event.data.length <= 3) return;
            if (!subscribed) {
                subscribed = true;
                return
            }
            list.push([new Date(), event.data]);
            if (list.length > 15) {
                list.shift();
            }
            html = "";
            for (let i = list.length - 1; i >= 0; i--) {
                html += get_date(list[i][0]) + " " + list[i][1] + "<br />";
            }
            document.getElementById("activity-feed").innerHTML = html;
        }
    }

    let generator_username = document.getElementById("generator-username");
    let generator_repository = document.getElementById("generator-repository");

    let generator_result_image = document.getElementById("generator-result-image");
    let generator_result_markdown = document.getElementById("generator-result-markdown");
    let generator_result_html = document.getElementById("generator-result-html");
    let image_url = "";
    let username = "webklex";
    let repository = "gohits";

    generator_username.onchange = generate;
    generator_repository.onchange = generate;

    function generate(){
        if (generator_username.value.length > 0) username = generator_username.value;
        if (generator_repository.value.length > 0) repository = generator_repository.value;

        image_url = domain + "/svg/" + username + "/" + repository;
        generator_result_image.setAttribute("src", image_url);
        generator_result_markdown.innerHTML = "[![Hits](" + image_url + ")](" + domain + ")";
        generator_result_html.innerHTML = '&lt;a href="' + domain + '">&lt;img src="' + image_url + '" alt="Hits"/>&lt;/a>';
    }

    connect_socket();
    generate();

})();