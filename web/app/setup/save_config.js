"use client";

import { useState, useEffect, useCallback, useRef } from "react";

export const SaveConfig = () => {
  const [config, setConfig] = useState();
  useEffect(() => {
    const run = async () => {
      const resp = await fetch("/config/");
      const data = await resp.json();
      setConfig(data);
    };
    run();
  }, []);

  const dialogRef = useRef();
  const submit = useCallback(
    async (evt) => {
      evt.preventDefault();
      const data = new FormData(evt.target);
      const config = data.get("config");
      const r = new Request("/load", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: config,
      });
      const resp = await fetch(r);
      await resp.text();
      dialogRef.current.showModal();
    },
    [dialogRef.current],
  );

  return (
    <div className="card shadow-lg shadow-primary m-4">
      <div className="card-body">
        <h3 className="card-title">Setup</h3>
        <div className="w-full">
          <form className="flex flex-col gap-4" onSubmit={submit}>
            <fieldset>
              <legend>Save config to caddy server</legend>
              {config != null ? (
                <textarea
                  name="config"
                  className="textarea textarea-primary w-full min-h-64"
                  placeholder="Put your config file here"
                  defaultValue={JSON.stringify(config, null, 2)}
                ></textarea>
              ) : (
                <div className="loading"></div>
              )}
            </fieldset>
            <div className="flex gap-4">
              <button className="btn btn-primary btn-sm rounded-xl">
                Setup
              </button>
            </div>
          </form>
          <dialog ref={dialogRef} className="modal">
            <div className="modal-box">
              <h3 className="text-lg font-bold text-success">Success!</h3>
              <p className="py-4">Config saved</p>
              <form method="dialog" className="modal-action">
                <button className="btn btn-primary btn-sm rounded-xl">
                  Close
                </button>
              </form>
            </div>
            <form method="dialog" className="modal-backdrop">
              <button>close</button>
            </form>
          </dialog>
        </div>
      </div>
    </div>
  );
};
