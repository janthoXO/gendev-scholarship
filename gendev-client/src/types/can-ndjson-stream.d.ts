declare module 'can-ndjson-stream' {
  /**
   * Transform a ReadableStream to parse newline delimited JSON
   * @param stream - The readable stream to transform
   * @returns A ReadableStream that emits parsed JSON objects
   */
    function ndjsonStream(stream: ReadableStream<Uint8Array> | null): ReadableStream<any>;

    export default ndjsonStream;

  // export default function ndjsonStream(data: unknown): {
  //   getReader: () => {
  //     read: () => Promise<{
  //       done: boolean;
  //       value: any;
  //     }>;
  //   };
  //   cancel: () => void;
  // };
}
