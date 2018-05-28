var Brush = function(lc, onPointerDownEvent, onPointerDragEvent, onPointerUpEvent, onPointerMoveEvent) {  // take lc as constructor arg
    var self = this;

    Brush.prototype.setNextBlot = function() {
      self.nextBlot=Math.floor(Math.random()*400);
    }
    self.blotFactor=0;
    self.setNextBlot();
    self.lastPt=null;

    Brush.prototype.vary = function(lc,pt) {
      var velocity=0,velocity_s=0;
      if (self.lastPt) {
        velocity=Math.abs((self.lastPt.x-pt.x)+(self.lastPt.y-pt.y))
        velocity_s=Math.sqrt(velocity)
        
      }
      self.nextBlot-=velocity_s;
      if (self.nextBlot <= 0)
      {
        if (velocity_s < 4) { velocity_s=1;}
        self.blotFactor=(velocity_s/4);
        self.setNextBlot();
      }
      pointSize=lc.tool.strokeWidth*(1+self.blotFactor);
      self.blotFactor*=0.8;
      self.lastPt=pt;
      return pointSize;
    };

    return {
      usesSimpleAPI: false,  // DO NOT FORGET THIS!!!
      name: 'Brush',
      iconName: 'line',
      strokeWidth: lc.opts.defaultStrokeWidth,
      optionsStyle: 'stroke-width',
      didBecomeActive: function(lc) {
        var onPointerDown = function(pt) {
          self.currentShape = LC.createShape('LinePath');
          self.currentShape.addPoint(LC.createShape('Point', { x: pt.x, y: pt.y, size: lc.tool.strokeWidth, color: lc.getColor('primary') }));
          lc.setShapesInProgress([self.currentShape]);
          onPointerDownEvent();
        };
  
        var onPointerDrag = function(pt) {
          pointSize=self.vary(lc,pt);
          
          self.currentShape.addPoint(LC.createShape('Point', { x: pt.x, y: pt.y, size: pointSize, color: lc.getColor('primary') }));
          lc.drawShapeInProgress(self.currentShape);
          onPointerDragEvent();
        };
  
        var onPointerUp = function(pt) {
          self.lastPt=null;
          lc.setShapesInProgress([]);
          lc.saveShape(self.currentShape);
          onPointerUpEvent();
        };
  
        var onPointerMove = function(pt) {
          onPointerMoveEvent();
          //console.log("Mouse moved to", pt);
        };
  
        // lc.on() returns a function that unsubscribes us. capture it.
        self.unsubscribeFuncs = [
          lc.on('lc-pointerdown', onPointerDown),
          lc.on('lc-pointerdrag', onPointerDrag),
          lc.on('lc-pointerup', onPointerUp),
          lc.on('lc-pointermove', onPointerMove),
          lc.on('setStrokeWidth', (function(_this) {
            return function(strokeWidth) {
              _this.strokeWidth = strokeWidth;
              return lc.trigger('toolDidUpdateOptions');
            };
          })(this))
        ];
      },
  
      willBecomeInactive: function(lc) {
        // call all the unsubscribe functions
        self.unsubscribeFuncs.map(function(f) { f() });
      }
    }
  };