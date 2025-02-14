import type { Meta, StoryObj } from '@storybook/react';
import Button from './Button';
import mockButtonProps from './Button.mocks';

// More on how to set up stories at: https://storybook.js.org/docs/7.0/react/writing-stories/introduction
const meta: Meta<typeof Button> = {
  title: 'Komiser/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {
    type: {
      control: {
        type: 'inline-radio'
      }
    },
    style: {
      control: {
        type: 'inline-radio'
      }
    },
    size: {
      control: {
        type: 'inline-radio'
      }
    }
  }
};

export default meta;
type Story = StoryObj<typeof Button>;

// More on writing stories with args: https://storybook.js.org/docs/7.0/react/writing-stories/args
export const Primary: Story = {
  args: {
    ...mockButtonProps.base
  }
};

export const Secondary: Story = {
  args: {
    ...mockButtonProps.secondary
  }
};

export const Bulk: Story = {
  args: {
    ...mockButtonProps.bulk
  }
};

export const BulkOutline: Story = {
  args: {
    ...mockButtonProps.bulkOutline
  }
};

export const Outline: Story = {
  args: {
    ...mockButtonProps.outline
  }
};

export const Ghost: Story = {
  args: {
    ...mockButtonProps.ghost
  }
};

export const Delete: Story = {
  args: {
    ...mockButtonProps.deleteButton
  }
};
export const DeleteGhost: Story = {
  args: {
    ...mockButtonProps.deleteButtonGhost
  }
};
