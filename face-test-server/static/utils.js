function imgPreview(fileDom, imgDom) {
	if (fileDom && fileDom.files[0]) {
		var reader = new FileReader();

		reader.onload = function(e) {
			imgDom.attr('src', e.target.result);
		};

		reader.readAsDataURL(fileDom.files[0]);
	}
}

function extractImageData(url) {
	return url.replace(/^data:image\/[^;]+;base64,/, '');
}

function htmlEscape(s) {
	return $("<div>").text(s).html();
}