(function(window) {
  window.env = window.env || {};

  // Environment variables
  window["env"]["production"] = false;
  window["env"]["apiUrl"] = "http://localhost:3030";
  window["env"]["debug"] = true;
})(this);