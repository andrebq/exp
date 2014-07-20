(function(){
    var cachedImages = {};

    function resizeMe(me, dimension) {
        var css = E.Css(me);
        css.setSize(dimension);
        css.set("display", "block");
    }

    function resizeBackdrop(bd, dimension) {
        var css = E.Css(bd);
        css.set("position", "absolute");
        css.set("margin", "0");
        css.setSize(dimension);
        css.set("background-color", "red");
    }

    function resizeContent(content, dimension) {
        var css = E.Css(content);
        var margin = 10;
        css.set("position", "absolute");
        css.set("max-width", dimension.width - margin);
        css.set("max-height", dimension.height - margin);
        css.set("width", "100%");
        css.set("height", "100%");

        // need to fetch the actual width / height of the 
        // content, in order to setup the margins
        css.set("transform", E.Css.translate(margin/2, margin/2));
    }

    Polymer('ed-blockui', {
        backdrop: "#backdrop",
        content: "#content",
        backgroundImage: "./assets/debut_dark/debut_dark.png",
        created: function() {
            this.$subs = new E.Rx.Util.SubManager();
            E.Css.loadDataUrl(this.resolvePath(this.backgroundImage), "image/png", cachedImages)
                .then(function(value){ this.$bg = value.data; });
        },
        $iAmVisible: function() {
            return this.$visible = false;
        },
        toggle: function(show) {
            this.$visible = show;
            if (show) {
                this.$resize(E.Rx.dimension(window));
                this.style.display = "block";
            } else {
                this.style.display = "none";
            }
        },
        $resize: function(dimension) {
            var nodes = this.getNodes();
            resizeMe(this, dimension);
            resizeBackdrop(nodes.backdrop, dimension)
            resizeContent(nodes.content, dimension);
        },
        attached: function() {
            this.$subs.add('resize', Rx.Observable.fromEvent(window, 'resize')
                .filter(this.$iAmVisible.bind(this))
                .map(E.Rx.dimension)
                .subscribe(this.$resize.bind(this)));
        },
        detached: function() {
            this.$subs.dispose();
        },
        getNodes: function() {
            return {
                backdrop: this.querySelector(this.backdrop),
                content: this.querySelector(this.content),
            }
        },
    });
}.bind(window)());
