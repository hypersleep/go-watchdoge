var watchDogeFrontend = angular.module('watchDogeFrontend', []);

watchDogeFrontend.controller('AppCtrl', function AppCtrl ($scope, $http) {
  $scope.getMetrics = function () {
    $http.get('http://localhost:8080/metrics?server=istream109').success( function (data) { $scope.data = data; } );
  };
  $scope.getMetrics();
});

watchDogeFrontend.directive('chart', function () {

  var margin = 20,
  width = 960,
  height = 500 - .5 - margin,
  color = d3.interpolateRgb("#f77", "#77f");

  return {
            restrict: 'E', // the directive can be invoked only by using <my-directive> tag in the template
            scope: { // attributes bound to the scope of the directive
              val: '='
            },
            link:function(scope, element, attrs) {

              var margin = {top: 20, right: 20, bottom: 30, left: 50},
                  width = 960 - margin.left - margin.right,
                  height = 500 - margin.top - margin.bottom;

              var parseDate = d3.time.format('%d-%B-%Y-%H-%M-%S').parse;             

              var x = d3.time.scale()
                  .range([0, width]);

              var y = d3.scale.linear()
                  .range([height, 0]);

              var xAxis = d3.svg.axis()
                  .scale(x)
                  .orient("bottom");

              var yAxis = d3.svg.axis()
                  .scale(y)
                  .orient("left");

              var area = d3.svg.area()
                  .x(function(d) { return x(d.date); })
                  .y0(height)
                  .y1(function(d) { return y(d.close); });

              var line = d3.svg.line()
                  .x(function(d) { return x(d.date); })
                  .y(function(d) { return y(d.close); });

              var svg = d3.select(element[0]).append("svg")
                  .attr("width", width + margin.left + margin.right)
                  .attr("height", height + margin.top + margin.bottom)
                .append("g")
                  .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

              scope.$watch('val', function (newVal, oldVal) {
                if (!newVal) {
                  return;
                }
                console.log(newVal);

                data = newVal

                data.forEach(function(d) {
                  d.date = parseDate(d.date);
                  d.close = +d.close;
                });

                data.sort(function(a, b){ return d3.ascending(a.date, b.date); });

                x.domain(d3.extent(data, function(d) { return d.date; }));
                y.domain([0, d3.max(data, function(d) { return d.close; })]);
                svg.append("g")
                  .attr("class", "x axis")
                  .attr("transform", "translate(0," + height + ")")
                  .call(xAxis);

                svg.append("g")
                    .attr("class", "y axis")
                    .call(yAxis)
                  .append("text")
                    .attr("transform", "rotate(-90)")
                    .attr("y", 6)
                    .attr("dy", ".71em")
                    .style("text-anchor", "end")
                    .text("Resident memory, kB");

                svg.append("path")
                  .datum(data)
                  .attr("class", "area")
                  .attr("d", area);
              });
            }
        }
});

