ymaps.ready(init);

var myMap,
    myPlacemark;

function init() {
    var cord = $('#coords').text() // адрес           

    var myGeocoder = ymaps.geocode(cord); // пытаюсь передать переменную 
    myGeocoder.then(
        function(res) {
            /*myMap.geoObjects.add(res.geoObjects);*/

            var adres = res.geoObjects.get(0)._geoObjectComponent._geometry._coordinates; // записываю координаты в переменную
          console.log(adres);

            myPlacemark = new ymaps.Placemark(adres, { // пытаюсь передать координаты и поставить метку 
               
            });

    myMap = new ymaps.Map("map", {
        center: adres,
        zoom: 12
    });

            myMap.geoObjects.add(myPlacemark);
        },
        function(err) {
          console.log(err)
            // обработка ошибки
        }
    );




}