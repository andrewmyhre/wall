zoom_level=1;
$(document).ready(function() {
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