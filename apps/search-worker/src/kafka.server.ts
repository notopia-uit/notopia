import { Logger } from '@nestjs/common';
import { KafkaOptions, ServerKafka } from '@nestjs/microservices';

import { KafkaConsumerState } from './health';

export class TrackedKafkaServer extends ServerKafka {
  private readonly retryLogger = new Logger(TrackedKafkaServer.name);
  private retryTimer?: NodeJS.Timeout;
  private stopping = false;

  constructor(
    options: Required<KafkaOptions>['options'],
    private readonly state: KafkaConsumerState
  ) {
    super(options);
  }

  override listen(callback: (err?: unknown) => void): Promise<void> {
    const connect = () => {
      if (this.stopping) return;
      void super.listen((error) => {
        if (this.stopping) return;
        if (!error) {
          callback();
          return;
        }
        this.state.down();
        this.retryLogger.warn('Kafka startup failed; retrying');
        void super
          .close()
          .catch(() => {
            this.retryLogger.error('Failed to close Kafka after startup failure');
          })
          .finally(() => {
            if (!this.stopping) this.retryTimer = setTimeout(connect, 1000);
          });
      });
    };
    connect();
    return Promise.resolve();
  }

  override async close(): Promise<void> {
    this.stopping = true;
    clearTimeout(this.retryTimer);
    this.state.down();
    await super.close();
  }

  protected override registerConsumerEventListeners(): void {
    super.registerConsumerEventListeners();
    const consumer = this.consumer;
    if (!consumer) return;
    for (const event of [consumer.events.GROUP_JOIN, consumer.events.HEARTBEAT]) {
      consumer.on(event, () => this.state.heartbeat());
    }
    for (const event of [
      consumer.events.DISCONNECT,
      consumer.events.STOP,
      consumer.events.CRASH,
      consumer.events.REBALANCING,
      consumer.events.REQUEST_TIMEOUT,
    ]) {
      consumer.on(event, () => this.state.down());
    }
  }
}
