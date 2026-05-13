// CSP-safe bindings for up dashboard
(function () {
  document.addEventListener(
    "click",
    function (e) {
      var btn = e.target.closest("[data-action]");
      if (btn) {
        var action = btn.getAttribute("data-action");
        if (action === "copy-short-url") {
          e.preventDefault();
          var payload = btn.getAttribute("data-copy");
          navigator.clipboard
            ?.writeText(payload)
            .then(function () {
              var orig = btn.textContent;
              btn.textContent = "COPIED!";
              setTimeout(function () {
                btn.textContent = orig;
              }, 1500);
            })
            .catch(() => {});
          return;
        }
      }

    },
    false,
  );
})();
