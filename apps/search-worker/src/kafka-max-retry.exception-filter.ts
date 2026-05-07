import { ArgumentsHost, Catch, Logger } from '@nestjs/common';
import { BaseExceptionFilter } from '@nestjs/core';
import { KafkaContext } from '@nestjs/microservices';
import { type Producer } from 'kafkajs';

@Catch()
export class KafkaMaxRetryExceptionFilter extends BaseExceptionFilter {
  private readonly logger = new Logger(KafkaMaxRetryExceptionFilter.name);

  constructor(
    private readonly producer: Producer,
    private readonly maxRetries: number,
    private readonly skipHandler?: (message: any) => Promise<void>
  ) {
    super();
  }

  override async catch(exception: unknown, host: ArgumentsHost) {
    const kafkaContext = host.switchToRpc().getContext<KafkaContext>();
    const message = kafkaContext.getMessage();
    const currentRetryCount = this.getRetryCountFromContext(kafkaContext);

    if (currentRetryCount >= this.maxRetries) {
      this.logger.warn(
        `Max retries (${this.maxRetries}) exceeded for message: ${JSON.stringify(message)}`
      );

      if (this.skipHandler) {
        try {
          await this.skipHandler(message);
        } catch (err) {
          this.logger.error('Error in skipHandler:', err);
        }
      }

      try {
        await this.republishWithRetry(kafkaContext, currentRetryCount + 1);
        await this.commitOffset(kafkaContext);
      } catch (republishError) {
        this.logger.error('Failed to republish message for retry:', republishError);
        super.catch(exception, host);
      }
    }
  }

  private getRetryCountFromContext(context: KafkaContext): number {
    const headers = context.getMessage().headers || {};
    const retryHeader = headers['retry-count'];
    if (!retryHeader) {
      return 0;
    }
    const value = Buffer.isBuffer(retryHeader) ? retryHeader.toString() : String(retryHeader);
    return parseInt(value, 10) || 0;
  }

  private async republishWithRetry(context: KafkaContext, retryCount: number): Promise<void> {
    const topic = context.getTopic();
    const message = context.getMessage();

    await this.producer.send({
      topic,
      messages: [
        {
          key: message.key,
          value: message.value,
          headers: {
            ...message.headers,
            'retry-count': retryCount.toString(),
          },
        },
      ],
    });
  }

  private async commitOffset(context: KafkaContext): Promise<void> {
    const consumer = context.getConsumer();
    if (!consumer) {
      throw new Error('Consumer instance is not available from KafkaContext.');
    }

    const topic = context.getTopic();
    const partition = context.getPartition();
    const message = context.getMessage();
    const offset = message.offset;

    if (!topic || partition === undefined || offset === undefined) {
      throw new Error('Incomplete Kafka message context for committing offset.');
    }

    await consumer.commitOffsets([
      {
        topic,
        partition,
        offset: (Number(offset) + 1).toString(),
      },
    ]);
  }
}
