require.config({
	shim: {
		'jquery': {exports: '$'},
		'uikit': ['jquery'],
		'uikit-datepicker': ['uikit'],
		'uikit-slideshow': ['uikit'],
		'uikit-slideshow-fx': ['uikit'],
		'uikit-slideset': ['uikit'],
		'uikit-sticky': ['uikit'],
		'uikit-tooltip': ['uikit'],
		'uikit-parallax': ['uikit'],
		'uikit-lightbox': ['uikit'],
		'uikit-grid': ['uikit'],
		'uikit-autocomplete': ['uikit'],
		'uikit-search': ['uikit'],
		'wow':['jquery', 'uikit'],
		'bootstrap':['jquery', 'tether'],
		'template' : ['jquery', 'tether'],
		'offcanvas-menu' : ['jquery', 'tether'],
		'ie10-viewport-bug-workaround' :['jquery', 'tether']
	},
	paths: {
		'jquery': "http://animedom/assets/bower_components/jquery/dist/jquery",
		'uikit': "http://animedom/assets/bower_components/uikit/js/uikit",
		'uikit-datepicker': "http://animedom/assets/bower_components/uikit/js/components/datepicker",
		'uikit-slideshow': "http://animedom/assets/bower_components/uikit/js/components/slideshow",
		'uikit-slideshow-fx': "http://animedom/assets/bower_components/uikit/js/components/slideshow-fx",
		'uikit-slideset': "http://animedom/assets/bower_components/uikit/js/components/slideset",
		'uikit-sticky': "http://animedom/assets/bower_components/uikit/js/components/sticky",
		'uikit-tooltip': "http://animedom/assets/bower_components/uikit/js/components/tooltip",
		'uikit-parallax': "http://animedom/assets/bower_components/uikit/js/components/parallax",
		'uikit-lightbox': "http://animedom/assets/bower_components/uikit/js/components/lightbox",
		'uikit-grid': "http://animedom/assets/bower_components/uikit/js/components/grid",
		'uikit-autocomplete': "http://animedom/assets/bower_components/uikit/js/components/autocomplete",
		'uikit-search': "http://animedom/assets/bower_components/uikit/js/components/search",
		'wow': "http://animedom/assets/bower_components/wow/dist/wow",
		'offcanvas-menu': "http://animedom/assets/js/offcanvas-menu",
		'template': "http://animedom/assets/js/template",
		'tether': "http://animedom/assets/bower_components/tether/dist/js/tether",
		'bootstrap': "http://animedom/assets/bower_components/bootstrap/dist/js/bootstrap",
		'ie10-viewport-bug-workaround': "http://animedom/assets/js/ie10-viewport-bug-workarouns"
	},
	priority: [
		"jquery",
		"uikit",
		"tether",
		"bootstrap"
	]
});


define(['wow', 'jquery','uikit','uikit-datepicker','uikit-slideshow','uikit-slideshow-fx','uikit-slideset','uikit-sticky','uikit-tooltip','uikit-parallax','uikit-lightbox','uikit-grid','uikit-autocomplete','uikit-search', 'bootstrap','template','offcanvas-menu','ie10-viewport-bug-workaround'], function (wow) {
	wow.init();
});
