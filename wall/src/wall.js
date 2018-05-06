if (window.location.hostname == 'localhost') {
    api_host='http://localhost:38000'
    brick_host='http://localhost:30080'
    wall_host='http://localhost:38001'
} else {
    api_host='http://wall-api.andrew-myhre.com'
    brick_host='http://brick.andrew-myhre.com'
    wall_host='http://wall.andrew-myhre.com'
}

row_length=Math.floor(screen.width/240);
column_length=Math.floor(screen.height/135);

function buildBrickHash(bricks) {
    var bricksHash=new Object();
    $.each(bricks, function(i, brick) {
        bricksHash[brick.id]=brick;
    });
    return bricksHash;
}

function render(bricksHash) {
    var i=0;
    for (var y=1; y<=column_length; y++) {
        for (var x=1; x<=row_length; x++) {
            brick_id=x+","+y;
            brick=bricksHash[brick_id];
            var brick_element=$("#wall #"+brick_id)[0]
            if (brick_element == null)  {
                brick_element=$("<div class='brick' />");
                var a=$("<a id='"+brick_id+"' href='"+brick_host+"/#"+brick_id+"' />");
                
                brick_element.append(a);

                prev=$("#wall .brick").eq(i);
                if (prev != null) {
                    $("#wall").append(brick_element);
                } else {
                    prev.after(brick_element);
                }
            }
            if (brick != null) {
                $(brick_element.children()[0]).html("<img src='"+api_host+brick.url+"' width=200 />");
            }
            i++;
        }
    }
    
}

$(document).ready(function(){
    var wall=$("#wall");
    var bricks=null;
    $.getJSON(api_host+'/bricks', function(response) {
        bricks=response;
        bricksHash=buildBrickHash(bricks);
        render(bricksHash);
    })
})