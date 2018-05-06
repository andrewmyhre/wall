function buildBrickHash(bricks) {
    var bricksHash=new Object();
    $.each(bricks, function(i, brick) {
        bricksHash[brick.id]=brick;
    });
    return bricksHash;
}

function render(bricksHash) {
    var i=0;
    for (var y=1; y<10; y++) {
        for (var x=1; x<10; x++) {
            brick_id=x+","+y;
            brick=bricksHash[brick_id];
            var brick_element=$("#wall #"+brick_id)[0]
            if (brick_element == null)  {
                brick_element=$("<div class='brick' />");
                var a=$("<a id='"+brick_id+"' href='http://localhost:30080/#"+brick_id+"' />");
                
                brick_element.append(a);

                prev=$("#wall .brick").eq(i);
                if (prev != null) {
                    $("#wall").append(brick_element);
                } else {
                    prev.after(brick_element);
                }
            }
            if (brick != null) {
                $(brick_element.children()[0]).html("<img src='http://localhost:38000"+brick.url+"' width=200 />");
            }
            i++;
        }
    }
    
}

$(document).ready(function(){
    var wall=$("#wall");
    var bricks=null;
    $.getJSON('http://localhost:38000/bricks', function(response) {
        bricks=response;
        bricksHash=buildBrickHash(bricks);
        render(bricksHash);
    })
})