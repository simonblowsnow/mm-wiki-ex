//jQuery拖拽缩放组件  create by dongp
//version 1.0
;
(function(jQuery, window, document, undfined) {
    ///拖拽，缩放    dongp
    var DragZoom = function(ele, opt) {
        this.$element = ele;
        this.defaults = {
            minzoom: 1,
            maxzoom: 5,
            zoom: 1,
            speed: 0.7,
            scope: null,
            onWheelStart: null,
            onWheelEnd: null,
            onDragStart: null,
            onDragMove: null,
            onDragEnd: null
        };

        this.options = $.extend({}, this.defaults, opt);
    }
    DragZoom.prototype = {
        Init: function() {
            var self = this;
            //参数
            self.x = this.$element.offset().left;
            self.y = this.$element.offset().top;
            self.width = this.$element.width();
            self.height = this.$element.height();
            self.scale = 1;
            self.relX = 0;
            self.relY = 0;
            self.isMoved = false;

            //缩放
            self.$element.on('mouseout', function(e) {
                $("body").css('cursor', 'default');
                return false;
            }).on('mousewheel', function(e, delta) {
                var size = delta * self.options.speed;
                self.options.zoom = (self.options.zoom * 10 + delta) / 10;
                self.wheel(e, self);
                return false;
            }).on('mousedown', function(e) {
                $("body").css("cursor", "move")
                self.start(e, self);
                return false;
            }).on('mouseup', function(e) {
                $("body").css('cursor', 'default');
            });

            //拖拽
            $(document).on('mousemove', function(e) {
                // if (self.options.zoom > 1) {
                //     self.move(e, self);
                // }
                self.move(e, self);
                return false;
            }).on('mouseup', function(e) {
                self.end(e, self);
                return false;
            });
            return self.$element;
        },
        wheel: function(ev, self) {

            if (self.options.zoom >= self.options.minzoom && self.options.zoom <= self.options.maxzoom) {
                //缩放开始回调
                self.options.onWheelStart && typeof self.options.onWheelStart == 'function' ? self.options.onWheelStart() : null;

                var cursor_x = ev.pageX,
                    cursor_y = ev.pageY;

                var eleOffset = self.$element.offset();
                self.x = eleOffset.left;
                self.y = eleOffset.top;

                self.x = self.x - (cursor_x - self.x) * (self.options.zoom - self.scale) / self.scale;
                self.y = self.y - (cursor_y - self.y) * (self.options.zoom - self.scale) / self.scale;

                self.scale = self.options.zoom;

                self.$element.width(self.width * self.scale).height(self.height * self.scale);
                self.$element.offset({
                    top: self.y,
                    left: self.x
                });

                //缩放结束回调
                self.options.onWheelEnd && typeof self.options.onWheelEnd == 'function' ? self.options.onWheelEnd() : null;
            }
            self.options.zoom = self.options.zoom < self.options.minzoom ? self.options.minzoom :
                (self.options.zoom > self.options.maxzoom ? self.options.maxzoom : self.options.zoom);
        },
        start: function(ev, self) {
            self.isMoved = true;
            var selfOffset = self.$element.offset();
            self.relX = ev.clientX - selfOffset.left;
            self.relY = ev.clientY - selfOffset.top;

            //拖拽开始回调
            self.options.onDragStart ? self.options.onDragStart() : null;

        },
        move: function(ev, self) {

            if (self.isMoved) {
                self.y = ev.clientY - self.relY;
                self.x = ev.clientX - self.relX;

                self.$element.offset({
                    top: self.y,
                    left: self.x
                });

                // self.$element.animate({ top: self.y + 'px', left: self.x + 'px' });

                //拖拽移动回调
                self.options.onDragMove && typeof self.options.onDragMove == 'function' ? self.options.onDragMove() : null;

            }
        },
        end: function(ev, self) {
            self.isMoved = false;
            // $(document).off('mousemove').off('mouseup');

            //拖拽结束回调
            self.options.onDragEnd && typeof self.options.onDragEnd == 'function' ? self.options.onDragEnd() : null;
        }
    };

    var dragzoom;
    jQuery.fn.dragZoom = function(options) {
        dragzoom = new DragZoom(this, options);
        return dragzoom.Init();
    }
    jQuery.fn.dragZoomClear = function() {
        if (dragzoom) {
            dragzoom.options.zoom = 1;
            dragzoom.scale = 1;
        }
    }

})($, window, document, undefined);
