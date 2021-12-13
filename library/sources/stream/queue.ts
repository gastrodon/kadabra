import axios from "axios";

import Config from "../../types/config";
import { ConfigKind } from "../../types/config-kind";
import { StreamConfig } from "../../types/stream-kind";

// TODO fn: types ??
export interface Handle {
  push: (data: string) => Promise<void>;
  head: () => Promise<string | null>;
}

export const attach = async (
  config: Config,
  stream: StreamConfig,
): Promise<Handle> => {
  const name = stream.name.split("/")[1];
  const host = config.get(ConfigKind.ScytherHost) as string;
  const url = host + "/queues/" + name;

  axios({
    url: host + "/queues",
    method: "POST",
    data: JSON.stringify({ name }),
  }).catch(
    (it) => {
      if (it.response?.status === 409) {
        return;
      }

      throw it;
    },
  );

  return {
    push: async (data: string) => {
      await axios({ url, method: "PUT", data });
    },
    head: async () => {
      try {
        return (await axios({ url: url + "/head", method: "GET" }))
          .data
          .message;
      } catch (error: any) {
        if (error.response.data?.error === "no_message") {
          return null;
        }

        throw error;
      }
    },
  };
};