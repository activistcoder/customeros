import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';

import { observer } from 'mobx-react-lite';
import { useFeatureIsOn } from '@growthbook/growthbook-react';
import {
  Droppable,
  DragDropContext,
  OnDragEndResponder,
} from '@hello-pangea/dnd';

import { Button } from '@ui/form/Button/Button';
import { useStore } from '@shared/hooks/useStore';
import { Columns02 } from '@ui/media/icons/Columns02';
import { Menu, MenuList, MenuGroup, MenuButton } from '@ui/overlay/Menu/Menu';

import { ColumnItem, DraggableColumnItem } from './ColumnItem';
import {
  invoicesOptionsMap,
  renewalsOptionsMap,
  invoicesHelperTextMap,
  renewalsHelperTextMap,
  organizationsOptionsMap,
  organizationsHelperTextMap,
} from './columnOptions';

interface EditColumnsProps {
  type: 'invoices' | 'renewals' | 'organizations';
}

export const EditColumns = observer(({ type }: EditColumnsProps) => {
  const isFeatureEnabled = useFeatureIsOn('edit-columns');
  const { tableViewDefsStore } = useStore();
  const [searchParams] = useSearchParams();
  const preset = searchParams?.get('preset');

  const [optionsMap, helperTextMap] = useMemo(() => {
    return [
      type === 'invoices'
        ? invoicesOptionsMap
        : type === 'renewals'
        ? renewalsOptionsMap
        : organizationsOptionsMap,
      type === 'invoices'
        ? invoicesHelperTextMap
        : type === 'renewals'
        ? renewalsHelperTextMap
        : organizationsHelperTextMap,
    ];
  }, [type]);

  const tableViewDef = tableViewDefsStore.getById(preset ?? '0');

  const columns =
    tableViewDef?.value?.columns.map((c) => ({
      ...c,
      label: optionsMap[c.columnType],
      helperText: helperTextMap[c.columnType],
    })) ?? [];

  const handleDragEnd: OnDragEndResponder = (res) => {
    const sourceIndex = res.source.index;
    const destIndex = res?.destination?.index as number;
    const destination = res.destination;

    if (!destination) return;
    if (sourceIndex === destIndex) return;

    console.log('reorder column');
    tableViewDef?.reorderColumn(sourceIndex, destIndex);
  };

  if (!isFeatureEnabled) return null;

  return (
    <>
      <Menu
        onOpenChange={(open) => {
          if (!open) {
            console.log('order columns by visibility');
            tableViewDef?.orderColumnsByVisibility();
          }
        }}
      >
        <MenuButton asChild>
          <Button size='xs' leftIcon={<Columns02 />}>
            Edit columns
          </Button>
        </MenuButton>
        <DragDropContext onDragEnd={handleDragEnd}>
          <MenuList className='w-[350px]'>
            <ColumnItem
              isPinned
              noPointerEvents
              label={columns?.[0]?.label}
              visible={columns?.[0]?.visible}
              columnType={columns?.[0]?.columnType}
            />
            <Droppable
              key='active-columns'
              droppableId='active-columns'
              renderClone={(provided, snapshot, rubric) => {
                return (
                  <ColumnItem
                    provided={provided}
                    snapshot={snapshot}
                    helperText={columns[rubric.source.index].helperText}
                    columnType={columns[rubric.source.index].columnType}
                    visible={columns[rubric.source.index].visible}
                    onCheck={() => {
                      tableViewDef?.update((value) => {
                        value.columns[rubric.source.index].visible =
                          !value.columns[rubric.source.index].visible;

                        return value;
                      });
                    }}
                    label={columns[rubric.source.index].label}
                  />
                );
              }}
            >
              {(provided, { isDraggingOver }) => (
                <>
                  <MenuGroup
                    ref={provided.innerRef}
                    {...provided.droppableProps}
                  >
                    {columns.map(
                      (col, index) =>
                        index > 0 && (
                          <DraggableColumnItem
                            index={index}
                            label={col?.label}
                            visible={col?.visible}
                            helperText={col?.helperText}
                            noPointerEvents={isDraggingOver}
                            key={col?.columnType}
                            onCheck={() => {
                              tableViewDef?.update((value) => {
                                value.columns[index].visible =
                                  !value.columns[index].visible;

                                return value;
                              });
                            }}
                            columnType={col?.columnType}
                          />
                        ),
                    )}
                    {provided.placeholder}
                  </MenuGroup>
                </>
              )}
            </Droppable>
          </MenuList>
        </DragDropContext>
      </Menu>
    </>
  );
});