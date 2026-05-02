import { ArgumentsHost, Catch, Logger } from '@nestjs/common';
import { BaseExceptionFilter } from '@nestjs/core';
import { KafkaContext } from '@nestjs/microservices';

// TODO: need to recheck & replace? https://github.com/nestjs/docs.nestjs.com/pull/3246
@Catch()
export class KafkaMaxRetryExceptionFilter extends BaseExceptionFilter {
  private readonly logger = new Logger(KafkaMaxRetryExceptionFilter.name);

  constructor(
    private readonly maxRetries: number,
    private readonly skipHandler?: (message: any) => Promise<void>
  ) {
    super();
  }

  override catch(exception: unknown, host: ArgumentsHost) {
    const kafkaContext = host.switchToRpc().getContext<KafkaContext>();
    const message = kafkaContext.getMessage();
    const currentRetryCount = this.getRetryCountFromContext(kafkaContext);

    if (currentRetryCount >= this.maxRetries) {
      this.logger.warn(
        `Max retries (${this.maxRetries}) exceeded for message: ${JSON.stringify(message)}`
      );

      if (this.skipHandler) {
        this.skipHandler(message).catch((err) => {
          this.logger.error('Error in skipHandler:', err);
        });
      }

      this.commitOffset(kafkaContext).catch((commitError) => {
        this.logger.error('Failed to commit offset:', commitError);
      });
      return;
    }

    super.catch(exception, host);
  }

  private getRetryCountFromContext(context: KafkaContext): number {
    const headers = context.getMessage().headers || {};
    const retryHeader = headers['retryCount'] || headers['retry-count'];
    return retryHeader ? Number(retryHeader) : 0;
  }

  private async commitOffset(context: KafkaContext): Promise<void> {
    const consumer = context.getConsumer && context.getConsumer();
    if (!consumer) {
      throw new Error('Consumer instance is not available from KafkaContext.');
    }

    const topic = context.getTopic && context.getTopic();
    const partition = context.getPartition && context.getPartition();
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
