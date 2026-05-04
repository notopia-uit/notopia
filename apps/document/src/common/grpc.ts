// import { Timestamp } from '@notopia-uit/pb/google/protobuf/timestamp';

// NOTE: Because I'm too tired why tsgo doesn't work, but lsp works
export function protoTimestampToDate(timestamp: { seconds: number; nanos: number }): Date {
  return new Date(timestamp.seconds * 1000 + timestamp.nanos / 1e6);
}

export function isGrpcError(error: unknown): error is { code?: number } {
  return typeof error === 'object' && error !== null && 'code' in error;
}
