if (window.location.hostname == 'localhost') {
    api_host='http://localhost:38000'
    brick_host='http://localhost:30080'
    wall_host='http://localhost:38001'
} else {
    api_host='http://wall-api.andrew-myhre.com'
    brick_host='http://brick.andrew-myhre.com'
    wall_host='http://wall.andrew-myhre.com'
}

zoom_level=1;
$(document).ready(function() {
    if (!window.location.hash) {
        alert("No brick id provided");
        return;
    }
    brick_id=window.location.hash.substring(1);
    //var img = new Image()
    //img.crossOrigin='Anonymous';
    //img.src = api_host+'/bricks/'+brick_id;
    var lc = null;
    var tools;
    var strokeWidths;
    var colors;

    var setCurrentByName;
    var findByName;

    // the only LC-specific thing we have to do
    var containerOne = document.getElementsByClassName('literally')[0];

    console.log("devicePixelRatio:" + window.devicePixelRatio)

    var showLC = function() {
        lc = LC.init(containerOne, {
            snapshot: JSON.parse(localStorage.getItem('drawing-'+brick_id)),
            defaultStrokeWidth: 10,
            strokeWidths: [10, 20, 50],
            secondaryColor: 'transparent',
            imageSize: {width: 1200, height: 675}
        });

        base_image = new Image();
        base_image.crossOrigin='Anonymous';
        base_image.src = api_host+'/bricks/'+brick_id;
        base_image.onload = function(){
            var b=$('.background')[0];
            var ctx=b.getContext("2d");
            ctx.drawImage(base_image, 0, 0, 1200, 675);
        }
        var b=$('.background')[0];
        b.width=lc.width;
        b.height=lc.height;
        window.demoLC = lc;

        var save = function() {
            localStorage.setItem('drawing-'+brick_id, JSON.stringify(lc.getSnapshot()));
        }

        lc.on('drawingChange', save);
        lc.on('pan', save);
        lc.on('zoom', save);

        $("#open-image").click(function() {
            window.open(lc.getImage({
            scale: 1, margin: {top: 10, right: 10, bottom: 10, left: 10}
            }).toDataURL());
        });

        $("#change-size").click(function() {
            lc.setImageSize(null, 200);
        });

        $("#reset-size").click(function() {
            lc.setImageSize(null, null);
        });

        $("#clear-lc").click(function() {
            lc.clear();
        });
        $("#zoom-in").click(function() {
            lc.zoom(0.1)
            $('.background')[0].getContext("2d").scale(lc.scale,lc.scale)
        });
        $("#zoom-out").click(function() {
            lc.zoom(-0.1)
            $('.background')[0].getContext("2d").scale(lc.scale,lc.scale)
        });

        $("#publish-lc").click(function() {
            console.log('publish image');
            $('.background')[0].getContext("2d").drawImage(lc.getImage({scaleDownRetina:true}),0,0);
            $('.literally').hide();

            $.ajax({
                url: api_host+"/bricks/"+brick_id,
                type: "PUT",
                dataType: "json",
                data: JSON.stringify({
                    imagedata: $('.background')[0].toDataURL()
                }),
                processData: false,
                contentType: false,
                success: function(xhr, status) {
                    console.log("server returned " + status);
                    lc.clear();
                    save();
                    window.location.href = wall_host;
                },
                error: function(xhr, status) {
                    console.log("error occurred:" + status);
                },
                complete: function(xhr, status) {
                    console.log("server returned " + status);
                }
            });
        });

        // Set up our own tools...
        tools = [
            {
            name: 'pencil',
            el: document.getElementById('tool-pencil'),
            tool: new LC.tools.Pencil(lc)
            },{
            name: 'eraser',
            el: document.getElementById('tool-eraser'),
            tool: new LC.tools.Eraser(lc)
            },{
            name: 'text',
            el: document.getElementById('tool-text'),
            tool: new LC.tools.Text(lc)
            },{
            name: 'line',
            el: document.getElementById('tool-line'),
            tool: new LC.tools.Line(lc)
            },{
            name: 'arrow',
            el: document.getElementById('tool-arrow'),
            tool: function() {
                arrow = new LC.tools.Line(lc);
                arrow.hasEndArrow = true;
                return arrow;
            }()
            },{
            name: 'dashed',
            el: document.getElementById('tool-dashed'),
            tool: function() {
                dashed = new LC.tools.Line(lc);
                dashed.isDashed = true;
                return dashed;
            }()
            },{
            name: 'ellipse',
            el: document.getElementById('tool-ellipse'),
            tool: new LC.tools.Ellipse(lc)
            },{
            name: 'tool-rectangle',
            el: document.getElementById('tool-rectangle'),
            tool: new LC.tools.Rectangle(lc)
            },{
            name: 'tool-polygon',
            el: document.getElementById('tool-polygon'),
            tool: new LC.tools.Polygon(lc)
            },{
            name: 'tool-pan',
            el: document.getElementById('tool-pan'),
            tool: new LC.tools.Pan(lc)
            },{
            name: 'tool-select',
            el: document.getElementById('tool-select'),
            tool: new LC.tools.SelectShape(lc)
            }
        ];

        strokeWidths = [
            {
            name: 10,
            el: document.getElementById('sizeTool-1'),
            size: 10
            },{
            name: 20,
            el: document.getElementById('sizeTool-2'),
            size: 20
            },{
            name: 50,
            el: document.getElementById('sizeTool-3'),
            size: 50
            }
        ];

        colors = [
            {
            name: 'black',
            el: document.getElementById('colorTool-black'),
            color: '#000000'
            },{
            name: 'blue',
            el: document.getElementById('colorTool-blue'),
            color: '#0000ff'
            },{
            name: 'red',
            el: document.getElementById('colorTool-red'),
            color: '#ff0000'
            }
        ];

        setCurrentByName = function(ary, val) {
            ary.forEach(function(i) {
            $(i.el).toggleClass('current', (i.name == val));
            });
        };

        findByName = function(ary, val) {
            var vals;
            vals = ary.filter(function(v){
            return v.name == val;
            });
            if ( vals.length == 0 )
            return null;
            else
            return vals[0];
        };

        // Wire tools
        tools.forEach(function(t) {
            $(t.el).click(function() {
            var sw;

            lc.setTool(t.tool);
            setCurrentByName(tools, t.name);
            setCurrentByName(strokeWidths, t.tool.strokeWidth);
            $('#tools-sizes').toggleClass('disabled', (t.name == 'text'));
            });
        });
        setCurrentByName(tools, tools[0].name);

        // Wire Stroke Widths
        // NOTE: This will not work until the stroke width PR is merged...
        strokeWidths.forEach(function(sw) {
            $(sw.el).click(function() {
            lc.trigger('setStrokeWidth', sw.size);
            setCurrentByName(strokeWidths, sw.name);
            })
        })
        setCurrentByName(strokeWidths, strokeWidths[0].name);

        // Wire Colors
        colors.forEach(function(clr) {
            $(clr.el).click(function() {
            lc.setColor('primary', clr.color)
            setCurrentByName(colors, clr.name);
            })
        })
        setCurrentByName(colors, colors[0].name);

    };

    $(document).ready(function() {
    // disable scrolling on touch devices so we can actually draw
    $(document).bind('touchmove', function(e) {
        if (e.target === document.documentElement) {
        return e.preventDefault();
        }
    });
    showLC();
    });

    $('#hide-lc').click(function() {
    if (lc) {
        lc.teardown();
        lc = null;
    }
    });

    $('#show-lc').click(function() {
    if (!lc) { showLC(); }
    });
    //console.log('set height');
    //$('.fs-container').height('2000px');
    //set_canvas_height();
});

$( window ).resize(function() {
    //set_canvas_height();
});

function set_canvas_height() {
    aspect_ratio=$(document).width() / $(document).height();
    console.log('aspect ratio:'+aspect_ratio)
    new_height=Math.floor($('.fs-container').width()/aspect_ratio);
    console.log('resizing fs-container to '+new_height+'px');
    $('.fs-container').height(new_height+'px');
    //window.demoLC.setImageSize($('.fs-container').width(),new_height)
}