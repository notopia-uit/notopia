'use client';

import {
  NoteWorkspaceTreeFolder,
  NoteWorkspaceTreeNote,
  getWorkspaceTreeOptions,
  showTrashOptions,
  useCreateFolderMutation,
  useCreateNoteMutation,
  useMoveWorkspaceItemsMutation,
  useTrashWorkspaceItemsMutation,
} from '@notopia-uit/api-gen';
import { useRenameFolderMutation, useRenameNoteMutation } from '@notopia-uit/api-gen';
import { Button } from '@notopia-uit/ui/components/shadcn/button';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@notopia-uit/ui/components/shadcn/context-menu';
import { Input } from '@notopia-uit/ui/components/shadcn/input';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useWorkspaceEvents } from '@notopia-uit/ui/contexts/workspace-events-context';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { cn } from '@notopia-uit/ui/lib/shadcn/utils';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { ChevronRight, FilePlus, FolderPlus, Trash2 } from 'lucide-react';
import { useRouter } from 'next/navigation';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  ControlledTreeEnvironment,
  DraggingPosition,
  Tree,
  type TreeItem,
  type TreeItemIndex,
  type TreeRef,
  type TreeViewState,
} from 'react-complex-tree';
type TreeData = Record<TreeItemIndex, TreeItem<string>>;

const mapDtoTreeData = (rootFolder: NoteWorkspaceTreeFolder) => {
  const tree: TreeData = {};

  const traverse = (node: NoteWorkspaceTreeFolder | NoteWorkspaceTreeNote, isFolder: boolean) => {
    const { id, name } = node;

    const treeItem: TreeItem<string> = {
      index: id,
      data: name,
      isFolder: isFolder,
    };

    if (isFolder) {
      const folderNode = node as NoteWorkspaceTreeFolder;

      const childrenIds: TreeItemIndex[] = [
        ...(folderNode.children || []).map((f) => f.id),
        ...(folderNode.notes || []).map((n) => n.id),
      ];

      if (childrenIds.length > 0) {
        treeItem.children = childrenIds;
      }

      folderNode.children?.forEach((childFolder) => traverse(childFolder, true));

      folderNode.notes?.forEach((note) => traverse(note, false));
    }

    tree[id] = treeItem;
  };

  traverse(rootFolder, true);

  return {
    treeData: tree,
    rootId: rootFolder.id,
  };
};

type DisposableType = {
  dispose: () => void;
};

type EventEmitterOptionsType<T> = {
  logger?: (log: string, payload?: T) => void;
};

type EventHandlerType<T> = ((payload: T) => Promise<void> | void) | null | undefined;

type TreeDataProviderType<T> = {
  getTreeItem: (itemId: TreeItemIndex) => Promise<TreeItem<T>>;
  getTreeItems?: (itemIds: TreeItemIndex[]) => Promise<TreeItem<T>[]>;
  onRenameItem?: (item: TreeItem<T>, name: string) => Promise<void>;
  onDidChangeTreeData?: (listener: (changedItemIds: TreeItemIndex[]) => void) => DisposableType;
  onChangeItemChildren?: (itemId: TreeItemIndex, newChildren: TreeItemIndex[]) => Promise<void>;
};

class EventEmitter<EventPayload> {
  private handlerCount = 0;

  private handlers: Array<EventHandlerType<EventPayload>> = [];

  private options?: EventEmitterOptionsType<EventPayload>;

  constructor(options?: EventEmitterOptionsType<EventPayload>) {
    this.options = options;
  }

  public get numberOfHandlers() {
    return this.handlers.filter((h) => !!h).length;
  }

  public async emit(payload: EventPayload): Promise<void> {
    const promises: Array<Promise<void>> = [];

    this.options?.logger?.('emit', payload);

    for (const handler of this.handlers) {
      if (handler) {
        const res = handler(payload) as Promise<void>;
        if (typeof res?.then === 'function') {
          promises.push(res);
        }
      }
    }

    await Promise.all(promises);
  }

  public on(handler: EventHandlerType<EventPayload>): number {
    this.options?.logger?.('on');
    this.handlers.push(handler);

    return this.handlerCount++;
  }

  public off(handlerId: number) {
    this.delete(handlerId);
  }

  public delete(handlerId: number) {
    this.options?.logger?.('off');
    this.handlers[handlerId] = null;
  }
}

class TreeDataProvider<T> implements TreeDataProviderType<T> {
  private items: Record<TreeItemIndex, TreeItem<T>>;
  private setItemName?: (item: TreeItem<T>, newName: string) => TreeItem<T>;

  public readonly onDidChangeTreeDataEmitter = new EventEmitter<TreeItemIndex[]>();

  constructor(
    items: Record<TreeItemIndex, TreeItem<T>>,
    setItemName?: (item: TreeItem<T>, newName: string) => TreeItem<T>
  ) {
    this.items = items;
    this.setItemName = setItemName;
  }

  public async getTreeItem(itemId: TreeItemIndex) {
    const item = this.items[itemId];
    if (!item) {
      return {
        index: itemId,
        isFolder: false,
        data: `Unknown Item: ${itemId}` as T,
      };
    }
    return Promise.resolve(item);
  }

  public async onChangeItemChildren(itemId: TreeItemIndex, newChildren: TreeItemIndex[]) {
    if (this.items[itemId]) {
      this.items[itemId].children = newChildren;
      await this.onDidChangeTreeDataEmitter.emit([itemId]);
    }
  }

  public onDidChangeTreeData(listener: (changedItemIds: TreeItemIndex[]) => void) {
    const handlerId = this.onDidChangeTreeDataEmitter.on((payload) => listener(payload));
    return { dispose: () => this.onDidChangeTreeDataEmitter.off(handlerId) };
  }

  public async onRenameItem(item: TreeItem<T>, name: string) {
    if (this.setItemName) {
      this.items[item.index] = this.setItemName(item, name);
    }
    return Promise.resolve();
  }
}

const viewStateInitial: TreeViewState = {
  'tree-sample': {},
};

const TreeView: React.FC<{ currentWorkspaceId: string }> = ({ currentWorkspaceId }) => {
  const queryClient = useQueryClient();
  const { retry } = useQueryErrorHandler();

  const {
    data,
    isError: isGetWorkSpaceTreeError,
    error: getWorkspaceTreeError,
    isPending: isGettingWorkspaceTree,
  } = useQuery({
    ...getWorkspaceTreeOptions({
      path: { workspaceId: currentWorkspaceId },
    }),
    select: (data) => mapDtoTreeData(data),
  });

  const workspaceTreeData = data?.treeData;
  const rootId = data?.rootId ?? '';
  const router = useRouter();
  const tree = useRef<TreeRef>(null);

  const { showAlert } = useAlert();

  const { subscribe } = useWorkspaceEvents();

  useEffect(() => {
    const unsubscribe = subscribe((event) => {
      if (
        event.event === 'WorkspaceItemsUpdatedEvent'
      ) {
        queryClient.invalidateQueries({
          queryKey: getWorkspaceTreeOptions({
            path: { workspaceId: currentWorkspaceId },
          }).queryKey,
        });
      }
    });
    return unsubscribe;
  }, [subscribe, queryClient, currentWorkspaceId]);

  const { mutate: renameNote } = useRenameNoteMutation({
    onError: (error) => {
      showAlert({
        type: 'error',
        title: 'Rename Note Failed',
        message: `${error instanceof Error ? error.message : 'Unknown error'}`,
      });
    },
  });
  const { mutate: renameFolder } = useRenameFolderMutation({
    onError: (error) => {
      showAlert({
        type: 'error',
        title: 'Rename Folder Failed',
        message: `${error instanceof Error ? error.message : 'Unknown error'}`,
      });
    },
  });
  //TODO:: call API to create new item and update tree data with response
  const { mutate: createNote } = useCreateNoteMutation({
    onSuccess: (responses, variables) => {
      const newNoteId = responses.id;
      const parentId = variables.body.folderId;
      setItems((prevItems) => {
        const newItems = { ...prevItems };
        const parentItem = newItems[parentId];
        if (parentItem && parentItem.isFolder) {
          newItems[parentId] = {
            ...parentItem,
            children: [...(parentItem.children ?? []), newNoteId],
          };
        }
        newItems[newNoteId] = {
          index: newNoteId,
          data: variables.body.name,
          isFolder: false,
        };
        return newItems;
      });
    },
    onError: (error) => {
      showAlert({
        type: 'error',
        title: 'Create Note Failed',
        message: `${error instanceof Error ? error.message : 'Unknown error'}`,
      });
    },
  });
  const { mutate: createFolder } = useCreateFolderMutation({
    onError: (error) => {
      showAlert({
        type: 'error',
        title: 'Create Folder Failed',
        message: `${error instanceof Error ? error.message : 'Unknown error'}`,
      });
    },
    onSuccess: (responses, variables) => {
      const newFolderId = responses.id;
      const parentId = variables.body.parentId ? variables.body.parentId : rootId;
      setItems((prevItems) => {
        const newItems = { ...prevItems };
        const parentItem = newItems[parentId];
        if (parentItem && parentItem.isFolder) {
          newItems[parentId] = {
            ...parentItem,

            children: [...(parentItem.children ?? []), newFolderId],
          };
        }
        newItems[newFolderId] = {
          index: newFolderId,
          data: variables.body.name,
          isFolder: true,
        };
        return newItems;
      });
    },
  });
  const [viewState, setViewState] = useState<TreeViewState>(viewStateInitial);
  const [search, setSearch] = useState<string | undefined>('');
  const [items, setItems] = useState<Record<TreeItemIndex, TreeItem<string>>>({});
  useEffect(() => {
    if (workspaceTreeData) {
      setItems(workspaceTreeData);
    }
  }, [workspaceTreeData]);

  const dataProvider = useMemo(() => new TreeDataProvider<string>(items), [items]);
  const getTargetParentId = useCallback(() => {
    const focusedId = viewState['tree-sample']?.focusedItem;
    let parentId: TreeItemIndex = rootId;

    if (focusedId && items[focusedId]) {
      if (items[focusedId].isFolder) {
        parentId = focusedId;
      } else {
        const parent = Object.values(items).find(
          (p) => p.isFolder && p.children?.includes(focusedId)
        );
        if (parent) parentId = parent.index;
      }
    }
    return parentId;
  }, [viewState, items, rootId]);

  const { mutateAsync: moveWorkspaceItems } = useMoveWorkspaceItemsMutation({
    onError: (error) => {
      showAlert({
        type: 'error',
        title: 'Move Items Failed',
        message: `${error instanceof Error ? error.message : 'Unknown error'}`,
      });
    },
  });

  const onDrop = useCallback(
    async (draggedItems: TreeItem<string>[], target: DraggingPosition) => {
      let destinationFolderId: TreeItemIndex = rootId;

      if (target.targetType === 'item') {
        const targetItem = items[target.targetItem];
        destinationFolderId = targetItem?.isFolder ? target.targetItem : rootId;
      } else if (target.targetType === 'between-items') {
        destinationFolderId = target.parentItem;
      } else if (target.targetType === 'root') {
        destinationFolderId = rootId;
      }

      const noteIds: string[] = [];
      const folderIds: string[] = [];

      for (const draggedItem of draggedItems) {
        if (items[draggedItem.index]?.isFolder) {
          folderIds.push(String(draggedItem.index));
        } else {
          noteIds.push(String(draggedItem.index));
        }
      }

      await moveWorkspaceItems({
        path: {
          workspaceId: currentWorkspaceId,
        },
        body: {
          noteIds,
          folderIds,
          destinationFolderId: String(destinationFolderId),
        },
      });

      setItems((prevItems) => {
        const newItems = { ...prevItems };

        for (const item of draggedItems) {
          const parent = Object.values(newItems).find(
            (p) => p.isFolder && p.children?.includes(item.index)
          );
          if (parent && parent.children) {
            newItems[parent.index] = {
              ...parent,
              children: parent.children.filter((child) => child !== item.index),
            };
          }
        }

        if (target.targetType === 'item') {
          const targetItem = newItems[target.targetItem];
          if (targetItem && targetItem.isFolder) {
            newItems[target.targetItem] = {
              ...targetItem,
              children: [...(targetItem.children || []), ...draggedItems.map((i) => i.index)],
            };
          }
        } else if (target.targetType === 'between-items') {
          const parentItem = newItems[target.parentItem];
          if (parentItem && parentItem.isFolder) {
            const newChildren = [...(parentItem.children || [])];
            newChildren.splice(target.childIndex, 0, ...draggedItems.map((i) => i.index));
            newItems[target.parentItem] = {
              ...parentItem,
              children: newChildren,
            };
          }
        } else if (target.targetType === 'root') {
          const rootItem = newItems[rootId];
          if (rootItem) {
            newItems[rootId] = {
              ...rootItem,
              children: [...(rootItem.children || []), ...draggedItems.map((i) => i.index)],
            };
          }
        }

        return newItems;
      });
    },
    [rootId, items, currentWorkspaceId, moveWorkspaceItems, setItems]
  );

  const { mutate: trashItems } = useTrashWorkspaceItemsMutation({
    onError: (error) => {
      showAlert({
        type: 'error',
        title: 'Trash Items Failed',
        message: `${error instanceof Error ? error.message : 'Unknown error'}`,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: showTrashOptions({ path: { workspaceId: currentWorkspaceId } }).queryKey,
      });
      showAlert({
        type: 'success',
        title: 'Items Trashed',
        message: 'Items have been moved to trash successfully',
      });
    },
  });

  const handleTrashItem = useCallback(
    (itemId: TreeItemIndex) => {
      const item = items[itemId];
      if (!item) return;

      const isFolder = item.isFolder;

      setItems((prevItems) => {
        const newItems = { ...prevItems };

        const parent = Object.values(newItems).find(
          (p) => p.isFolder && p.children?.includes(itemId)
        );
        if (parent && parent.children) {
          newItems[parent.index] = {
            ...parent,
            children: parent.children.filter((child) => child !== itemId),
          };
        }

        delete newItems[itemId];
        return newItems;
      });

      trashItems({
        path: {
          workspaceId: currentWorkspaceId,
        },
        body: {
          noteIds: isFolder ? [] : [String(itemId)],
          folderIds: isFolder ? [String(itemId)] : [],
        },
      });
    },
    [items, currentWorkspaceId, trashItems]
  );

  const handleCreateItem = useCallback(
    (isFolder: boolean) => {
      const parentId = getTargetParentId();
      const defaultName = isFolder ? 'New Folder' : 'New Note';

      if (isFolder) {
        createFolder({
          body: {
            workspaceId: currentWorkspaceId,
            icon: '📁',
            name: defaultName,
            parentId: parentId as string,
          },
        });
      } else {
        createNote({
          body: {
            folderId: parentId as string,
            icon: '📝',
            name: defaultName,
          },
        });
      }
    },
    [getTargetParentId, createFolder, createNote, currentWorkspaceId]
  );

  const getItemPath = useCallback(
    async (search: string, searchRoot: TreeItemIndex = rootId): Promise<TreeItemIndex[] | null> => {
      const searchTree = async (currentId: TreeItemIndex): Promise<TreeItemIndex[] | null> => {
        const item = await dataProvider.getTreeItem(currentId);

        if (item.data.toLowerCase().includes(search.toLowerCase())) {
          return [item.index];
        }

        if (item.children && item.children.length > 0) {
          const results = await Promise.all(item.children.map((childId) => searchTree(childId)));

          const foundPath = results.find((path): path is TreeItemIndex[] => path !== null);

          if (foundPath) {
            return [item.index, ...foundPath];
          }
        }

        return null;
      };

      return searchTree(searchRoot);
    },
    [dataProvider, rootId]
  );

  const onSubmit = useCallback(
    (e: React.SubmitEvent<HTMLFormElement>) => {
      e.preventDefault();
      if (search) {
        getItemPath(search)
          .then((path) => {
            if (path) {
              return tree.current?.expandSubsequently(path).then(() => {
                tree.current?.selectItems([path.at(-1) ?? '']);
                tree.current?.focusItem(path.at(-1) ?? '');
                tree.current?.toggleItemSelectStatus(path.at(-1) ?? '');
              });
            }
            return;
          })
          .catch((error) => {
            console.error('Error getting item:', error);
          });
      }
    },
    [getItemPath, search]
  );

  if (isGetWorkSpaceTreeError) {
    return (
      <div className="flex h-full items-center justify-center p-4">
        <QueryErrorFallback
          error={getWorkspaceTreeError}
          onRetry={retry}
          title="Failed to Load Workspace Tree"
          description="Unable to load your workspace folder and file structure. Please try again."
        />
      </div>
    );
  }

  //TODO: maybe use skeleton?
  return isGettingWorkspaceTree ? (
    <Spinner />
  ) : (
    <div className="flex size-full flex-col gap-4 overflow-hidden">
      <div className="flex shrink-0 flex-col gap-2">
        <form onSubmit={onSubmit} className="flex items-center gap-2">
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search..."
            className="h-8 text-sm"
          />
          <Button type="submit">Search</Button>
        </form>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            size-3="true"
            className="h-8 flex-1 text-xs" // flex-1 makes them share the space equally
            onClick={() => handleCreateItem(false)}
          >
            <FilePlus className="mr-1.5 size-3" />
            New Note
          </Button>
          <Button
            variant="outline"
            size="sm"
            size-3="true"
            className="h-8 flex-1 text-xs"
            onClick={() => handleCreateItem(true)}
          >
            <FolderPlus className="mr-1.5 size-3" />
            New Folder
          </Button>
        </div>
      </div>
      <div className="flex-1 overflow-auto">
        <ControlledTreeEnvironment<string>
          items={items}
          getItemTitle={(item) => item.data}
          canSearch={false}
          canSearchByStartingTyping={false}
          canDragAndDrop={true}
          canReorderItems={true}
          canDropOnFolder={true}
          canRename={true}
          onRenameItem={(item, name) => {
            if (item.isFolder) {
              renameFolder({
                path: {
                  folderId: item.index as string,
                },
                body: {
                  name: name,
                },
              });
              return;
            }
            renameNote({
              path: {
                noteId: item.index as string,
              },
              body: {
                name: name,
              },
            });
          }}
          viewState={viewState}
          onDrop={onDrop}
          onExpandItem={(item, treeId) => {
            setViewState((prevViewState) => ({
              ...prevViewState,
              [treeId]: {
                ...prevViewState[treeId],
                expandedItems: [...(prevViewState[treeId]?.expandedItems ?? []), item.index],
              },
            }));
          }}
          onCollapseItem={(item, treeId) => {
            setViewState((prevViewState) => ({
              ...prevViewState,
              [treeId]: {
                ...prevViewState[treeId],
                expandedItems:
                  prevViewState[treeId]?.expandedItems?.filter((id) => id !== item.index) ?? [],
              },
            }));
          }}
          onFocusItem={(item, treeId) => {
            setViewState((prevViewState) => ({
              ...prevViewState,
              [treeId]: {
                ...prevViewState[treeId],
                focusedItem: item.index,
              },
            }));
          }}
          onSelectItems={(selectedItems, treeId) => {
            const selectedId = selectedItems.at(-1) ?? '';
            setViewState((prevViewState) => ({
              ...prevViewState,
              [treeId]: {
                ...prevViewState[treeId],
                selectedItems: [selectedId],
              },
            }));
            if (selectedId && items[selectedId] && !items[selectedId].isFolder) {
              router.push(`/workspace/${currentWorkspaceId}/note/${selectedId}`);
            }
          }}
          renderTreeContainer={({ children, containerProps }) => {
            return (
              <div {...containerProps} className="rounded-md border">
                {children}
              </div>
            );
          }}
          renderLiveDescriptorContainer={({}) => <></>}
          renderItemsContainer={({ children, containerProps }) => {
            return <ul {...containerProps}>{children}</ul>;
          }}
          renderItem={({ title, item, arrow, context, depth, children }) => {
            const indentation = 10 * depth;
            return (
              <li
                {...context.itemContainerWithChildrenProps}
                className="my-1 [&>button]:aria-selected:bg-primary/15 [&>button>svg]:aria-expanded:rotate-90"
              >
                <ContextMenu>
                  <ContextMenuTrigger asChild>
                    <Button
                      {...context.itemContainerWithoutChildrenProps}
                      {...context.interactiveElementProps}
                      type="button"
                      variant="outline"
                      size="sm"
                      className={cn(
                        `grid h-8 w-full grid-flow-col items-center justify-start gap-2 border-none px-2 text-xs shadow-none`,
                        'focus:bg-secondary/20 hover:bg-secondary/30 transition-colors'
                      )}
                      style={{
                        paddingLeft: `${item.isFolder ? indentation : indentation + 16}px`,
                      }}
                    >
                      {item.isFolder && arrow}
                      {title}
                    </Button>
                  </ContextMenuTrigger>
                  <ContextMenuContent>
                    <ContextMenuItem
                      variant="destructive"
                      onClick={() => handleTrashItem(item.index)}
                    >
                      <Trash2 className="mr-2 size-4" />
                      Move to Trash
                    </ContextMenuItem>
                  </ContextMenuContent>
                </ContextMenu>
                {children}
              </li>
            );
          }}
          renderItemArrow={({ context }) => {
            return <ChevronRight {...context.arrowProps} className="size-3.5 shrink-0 text-muted-foreground!" />;
          }}
          renderItemTitle={({ title }) => <span className="truncate">{title}</span>}
        >
          <Tree ref={tree} treeId="tree-sample" rootItem={rootId} treeLabel="Sample Tree" />
        </ControlledTreeEnvironment>
      </div>
    </div>
  );
};

export default TreeView;
