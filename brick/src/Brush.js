var Brush = function(lc) {  // take lc as constructor arg
    var self = this;

    Brush.prototype.setNextBlot = function() {
      self.nextBlot=Math.floor(Math.random()*400);
    }
    self.blotFactor=0;
    self.setNextBlot();

    Brush.prototype.vary = function(lc) {
      self.nextBlot--;
      if (self.nextBlot <= 0)
      {
        self.blotFactor=(1+Math.random()*4);
        self.setNextBlot();
      }
      pointSize=lc.tool.strokeWidth*(0.3+Math.random()*0.7)*(1+self.blotFactor);
      self.blotFactor*=0.8;
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
          self.currentShape.addPoint(LC.createShape('Point', { x: pt.x, y: pt.y, size: lc.tool.strokeWidth, color: this.color }));
          lc.setShapesInProgress([self.currentShape]);
        };
  
        var onPointerDrag = function(pt) {
          pointSize=self.vary(lc);
          
          self.currentShape.addPoint(LC.createShape('Point', { x: pt.x, y: pt.y, size: pointSize, color: this.color }));
          lc.drawShapeInProgress(self.currentShape);
        };
  
        var onPointerUp = function(pt) {
          lc.setShapesInProgress([]);
          lc.saveShape(self.currentShape);
        };
  
        var onPointerMove = function(pt) {
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