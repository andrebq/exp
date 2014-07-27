(function(){
    Polymer('ed-sidebar', {
        rootFolder:"./",
        searchString: "",
        showItems: function(list) {
            this.$.navigation.items = list;
        },
        show: function() {
            this.$.searchbox.clear();
            this.$.searchbox.focus();
        },
        watchForChanges: function(onlyRemove) {
            if (this.subscription) {
                this.subscription.dispose();
            }
            if (onlyRemove) { return; }
            var that = this;
            this.subscription = Rx.Observable.fromEvent(this.$.searchbox, "change")
                .map(function(e){
                    return e.detail.value;
                })
                .map(this.fuzzySet.filter.bind(this.fuzzySet))
                .subscribe(function(list){
                    that.showItems(list);
                });
        },
        mergeContents: function(ev){
            var files = S(this.$.serverContents.response).trim().split("\n");
            this.fileList.merge(files);
            this.$.treelist.tree = this.fileList.tree();
            this.fuzzySet.resetItems(this.fileList.items());
        },
        attached: function(){
            this.watchForChanges();
        },
        detached: function() {
            this.watchForChanges(true);
        },
        ready: function(){
            this.fileList = new E.Fs.Filelist();
            this.fuzzySet = new E.Fs.Fuzzyset();
            this.$.serverContents.go();
        },
    });
}());
