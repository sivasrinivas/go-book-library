var libraryApp = angular.module('libraryApp', ['ngCookies']);

libraryApp.controller('libraryController', function libraryController($scope, $http, $cookies, $filter) {
    $scope.books = null;
    $scope.searchBooks = null;
    $scope.libraryPage = true;
    $scope.searchPage = false;
    $scope.reverseSort = false;
    $scope.orderByField = null;

    //On page load, get all added books
    $http.get("/books").then(function (response) {
        $scope.books = response.data;
        $scope.orderByField = $cookies.get('orderBy');
    });

    $scope.setCookie = function (orderBy) {
        $cookies.put('orderBy', orderBy);
    };

    $scope.showLibraryPage = function () {
        $scope.libraryPage = true;
        $scope.searchPage = false;
    };

    $scope.showSearchPage = function () {
        $scope.libraryPage = false;
        $scope.searchPage = true;
    };

    $scope.getBooks = function () {
        //Get books list
        $http.get("/search?search-string=" + $scope.searchStr).then(function (response) {
            $scope.searchBooks = response.data;
        });
    };

    $scope.addBook = function (book) {
        $http.post("/books/" + book.Id).then(function (response) {
            $scope.books.push(response.data)
        });
    };

    $scope.deleteBook = function (index, pk) {
        $http.delete("books/" + pk).then(function (response) {
            if (response.status == 200) {
                $scope.books = $filter('filter')($scope.books, function (book) {
                    return book.pk != pk
                });
            }
        })
    }
});