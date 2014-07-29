(function(){
    Polymer('ed-sidebar', {
        rootFolder:"./",
        searchString: "",
        $toggleTreelist: function(show) {
            if (show) {
                E.Css(this.$.treelist).restore('display');
            } else {
                E.Css(this.$.treelist).setWithBackup('display', 'none');
            }
        },
        showItems: function(list) {
            this.$.navigation.items = list;
        },
        created: function() {
            this.$subs = new E.Rx.Util.SubManager();
        },
        show: function() {
            this.$.searchbox.clear();
            this.$.searchbox.focus();
        },
        mergeContents: function(ev){
            var files = S(this.$.serverContents.response).trim().split("\n");
            this.fileList.merge(files);
            this.$.treelist.tree = this.fileList.tree();
            this.fuzzySet.resetItems(this.fileList.items());
        },
        attached: function(){
            this.$subs.add('searchbox-change', Rx.Observable.fromEvent(this.$.searchbox, "change")
                .pluck('detail').pluck('value')
                .filter(E.Rx.asBoolean)
                .map(E.Rx.exec(function(val) {
                    this.$toggleTreelist(!E.Rx.asBoolean(val));
                }.bind(this)))
                .map(this.fuzzySet.filter.bind(this.fuzzySet))
                .subscribe(function(list){
                    this.showItems(list);
                }.bind(this)));
            this.$subs.add('search-completed', Rx.Observable.fromEvent(this.$.searchbox, 'search-completed')
                .pluck('detail').pluck('value')
                .subscribe(function(value){
                    var items = this.fuzzySet.filter(value);
                    if (items.length > 0) {
                        this.fire('open-file', { filename: items[0] });
                    }
                }.bind(this)));
            this.$subs.add("create-new-file", Rx.Observable
                .fromEvent(this.$.treelist, "create-new-file")
                .subscribe(console.log.bind(console)));
        },
        detached: function() {
            this.$subs.dispose();
        },
        ready: function(){
            this.fileList = new E.Fs.Filelist();
            this.fuzzySet = new E.Fs.Fuzzyset();
            this.$.serverContents.go();
        },
    });
}());
