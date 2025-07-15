"use client";

import { useState, useEffect } from "react";

export const Config = ({ dataPromise }) => {
  const [data, setData] = useState({});
  useEffect(() => {
    fetch("/config/")
      .then((resp) => resp.json())
      .then((data) => setData(data));
  }, []);

  return <pre className="overflow-auto">{JSON.stringify(data, null, 2)}</pre>;
};
