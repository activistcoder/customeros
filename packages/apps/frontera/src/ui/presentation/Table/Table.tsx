import type {
  ColumnDef,
  OnChangeFn,
  SortingState,
  RowSelectionState,
  ColumnFiltersState,
  Table as TableInstance,
} from '@tanstack/react-table';

import React, {
  useRef,
  useMemo,
  useState,
  useEffect,
  forwardRef,
  HTMLAttributes,
  MutableRefObject,
} from 'react';

import { twMerge } from 'tailwind-merge';
import { useVirtualizer } from '@tanstack/react-virtual';
import {
  createRow,
  flexRender,
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFacetedRowModel,
  getFilteredRowModel,
} from '@tanstack/react-table';

import { cn } from '@ui/utils/cn';
import { Tumbleweed } from '@ui/media/icons/Tumbleweed';
import { Checkbox, CheckboxProps } from '@ui/form/Checkbox/Checkbox';

declare module '@tanstack/table-core' {
  // REASON: TData & TValue are not used in this interface but need to be defined
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnDefBase<TData, TValue = unknown> {
    fixWidth?: boolean;
    maxW?: number | string;
    skeleton: () => React.ReactNode;
  }
}

interface TableProps<T extends object> {
  data: T[];
  rowHeight?: number;
  isLoading?: boolean;
  totalItems?: number;
  borderColor?: string;
  sorting?: SortingState;
  canFetchMore?: boolean;
  onFetchMore?: () => void;
  fullRowSelection?: boolean;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  columns: ColumnDef<T, any>[];
  enableRowSelection?: boolean;
  enableTableActions?: boolean;
  contentHeight?: number | string;
  onFullRowSelection?: (id?: string) => void;
  onSortingChange?: OnChangeFn<SortingState>;
  tableRef: MutableRefObject<TableInstance<T> | null>;
  // REASON: Typing TValue is too exhaustive and has no benefit
  renderTableActions?: (table: TableInstance<T>) => React.ReactNode;
}

export const Table = <T extends object>({
  data,
  columns,
  tableRef,
  isLoading,
  onFetchMore,
  canFetchMore,
  totalItems = 40,
  onSortingChange,
  sorting: _sorting,
  renderTableActions,
  enableRowSelection,
  enableTableActions,
  fullRowSelection,
  rowHeight = 64,
  contentHeight,
  borderColor,
  onFullRowSelection,
}: TableProps<T>) => {
  const scrollElementRef = useRef<HTMLDivElement>(null);
  const [sorting, setSorting] = useState<SortingState>([]);
  const table = useReactTable<T>({
    data,
    columns,
    state: {
      sorting: _sorting ?? sorting,
    },
    manualSorting: true,
    enableRowSelection: enableRowSelection || fullRowSelection,
    enableMultiRowSelection: enableRowSelection && !fullRowSelection,
    enableColumnFilters: true,
    enableSortingRemoval: false,
    getCoreRowModel: getCoreRowModel<T>(),
    getSortedRowModel: getSortedRowModel<T>(),
    getFacetedRowModel: getFacetedRowModel<T>(),
    getFilteredRowModel: getFilteredRowModel<T>(),
    onSortingChange: onSortingChange ?? setSorting,
  });

  const { rows } = table.getRowModel();
  const rowVirtualizer = useVirtualizer({
    count: !data.length && isLoading ? 40 : totalItems,
    overscan: 30,
    getScrollElement: () => scrollElementRef.current,
    estimateSize: () => rowHeight,
  });

  const virtualRows = rowVirtualizer.getVirtualItems();

  useEffect(() => {
    const [lastItem] = [...virtualRows].reverse();

    if (!lastItem) {
      return;
    }

    if (lastItem.index >= data.length - 1 && canFetchMore && !isLoading) {
      onFetchMore?.();
    }
  }, [
    onFetchMore,
    data.length,
    isLoading,
    totalItems,
    virtualRows,
    canFetchMore,
  ]);

  useEffect(() => {
    if (tableRef) {
      tableRef.current = table;
    }
  }, [table]);

  const skeletonRow = useMemo(
    () => createRow<T>(table, 'SKELETON', {} as T, totalItems + 1, 0),
    [table, totalItems],
  );
  const THeaderMinW =
    table.getCenterTotalSize() + (enableRowSelection ? 28 : 0);

  return (
    <div className='flex w-full flex-col relative'>
      <TContent
        ref={scrollElementRef}
        height={contentHeight}
        borderColor={borderColor}
      >
        <THeader className='top-0 sticky ' style={{ minWidth: THeaderMinW }}>
          {table.getHeaderGroups().map((headerGroup) => {
            const width = enableRowSelection ? 'w-7' : 'w-2';

            return (
              <THeaderGroup key={headerGroup.id}>
                <THeaderCell className={cn('p-0 min-h-8', width)} />
                {headerGroup.headers.map((header, index) => {
                  const minWidth = header.getSize();
                  const maxWidth = header.column.columnDef.fixWidth
                    ? `${header.getSize()}px`
                    : 'none';
                  const flex = header.colSpan ?? '1';
                  const paddingRight = index === 0 && 'pr-0';
                  const paddingLeft =
                    index === 0 ? 'pl-2' : index === 1 ? 'pl-0' : 'pl-6';

                  const isHidden = header.column.columnDef.enableHiding;

                  if (isHidden) return null;

                  return (
                    <THeaderCell
                      key={header.id}
                      className={cn(paddingRight, paddingLeft)}
                      style={{ minWidth, maxWidth, flex }}
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext(),
                          )}
                    </THeaderCell>
                  );
                })}
              </THeaderGroup>
            );
          })}
        </THeader>
        <TBody className='w-full '>
          {!virtualRows.length && !isLoading && <NoResults />}
          {virtualRows.map((virtualRow) => {
            const row = rows[virtualRow.index];
            const minH = `${virtualRow.size}px`;

            const minW =
              table.getCenterTotalSize() + (enableRowSelection ? 28 : 0);

            const top = `${virtualRow.start}px`;

            const backgroundColor =
              virtualRow.index % 2 === 0 ? 'bg-gray-25' : 'bg-white';

            const hoverStyle = fullRowSelection
              ? 'hover:cursor-pointer'
              : 'group';

            const enabledRowOpacity = enableRowSelection
              ? 'opacity-100'
              : 'opacity-0';

            const enabledRowPointer = enableRowSelection
              ? 'pointer-events-auto'
              : 'pointer-events-none';

            const fullRowSelectionStyleDynamic = cn(
              virtualRow.index === 0
                ? 'hover:before:top[-1px]'
                : 'hover:before:top-[-2px]',
              `
              hover:after:contents-[""] hover:after:h-[2px] hover:after:w-full hover:after:bg-gray-200 hover:after:bottom-[-1px] hover:after:absolute
              hover:before:contents-[""] hover:before:w-full  hover:before:bg-gray-200 hover:before:h-[2px] hover:before:absolute`,
            );

            const rowHoverStyle = fullRowSelection
              ? fullRowSelectionStyleDynamic
              : undefined;

            const selectedStyle =
              fullRowSelection &&
              cn(
                'data-[selected=true]:before:contents-[""] data-[selected=true]:before:h-[2px] data-[selected=true]:before:w-full data-[selected=true]:before:bg-gray-200 data-[selected=true]:before:absolute',
                'data-[selected=true]:after:contents-[""] data-[selected=true]:after:w-full data-[selected=true]:after:bottom-[-1px] data-[selected=true]:after:bg-gray-200 data-[selected=true]:after:h-[2px] data-[selected=true]:after:absolute',
                virtualRow.index === 0
                  ? 'data-[selected=true]:before:top[-1px]'
                  : 'data-[selected=true]:before:top-[-2px]',
              );

            return (
              <TRow
                className={twMerge(
                  backgroundColor,
                  hoverStyle,
                  rowHoverStyle,
                  selectedStyle,
                  'group',
                )}
                style={{
                  minHeight: minH,
                  minWidth: minW,
                  top: top,
                }}
                key={virtualRow.key}
                data-selected={row?.getIsSelected()}
                data-index={virtualRow.index}
                ref={rowVirtualizer.measureElement}
                onClick={
                  fullRowSelection
                    ? (s) => {
                        row?.getToggleSelectedHandler()(s);
                        /// @ts-expect-error improve this later
                        const rowId = (row.original as unknown)?.id;
                        onFullRowSelection?.(rowId);
                      }
                    : undefined
                }
              >
                <TCell className='pl-2 pr-0 max-w-fit'>
                  {!fullRowSelection && (
                    <div
                      className={cn(
                        enabledRowPointer,
                        enabledRowOpacity,
                        'items-center ',
                      )}
                    >
                      {enableRowSelection && (
                        <MemoizedCheckbox
                          className='group-hover:visible group-hover:opacity-100  '
                          key={`checkbox-${virtualRow.index}`}
                          isChecked={row?.getIsSelected()}
                          disabled={!row || !row?.getCanSelect()}
                          onChange={(isChecked) =>
                            row?.getToggleSelectedHandler()(isChecked)
                          }
                        />
                      )}
                    </div>
                  )}
                </TCell>
                {(row ?? skeletonRow).getAllCells()?.map((cell, index) => {
                  const paddingRight = index === 0 ? 'pr-0' : '';
                  const paddingLeft =
                    index === 0 ? 'pl-2' : index === 1 ? 'pl-0' : 'pl-6';
                  const minWidth = cell.column.getSize();
                  const maxWidth = cell.column.columnDef.fixWidth
                    ? cell.column.getSize()
                    : 'none';

                  const flex =
                    table
                      .getFlatHeaders()
                      .find((h) => h.id === cell.column.columnDef.id)
                      ?.colSpan ?? '1';

                  const isHidden = cell.column.columnDef.enableHiding;

                  if (isHidden) return null;

                  return (
                    <TCell
                      key={cell.id}
                      className={cn(paddingRight, paddingLeft)}
                      style={{ minWidth, maxWidth, flex }}
                      data-index={cell.row.index}
                    >
                      {row
                        ? flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext(),
                          )
                        : cell.column.columnDef?.skeleton?.()}
                    </TCell>
                  );
                })}
              </TRow>
            );
          })}
        </TBody>
      </TContent>

      {enableTableActions && <TActions>{renderTableActions?.(table)}</TActions>}
    </div>
  );
};

interface GenericProps {
  tabIndex?: number;
  className?: string;
  children?: React.ReactNode;
  style?: React.CSSProperties;
  onClick?: (event: React.MouseEvent<HTMLDivElement, MouseEvent>) => void;
}

const TBody = forwardRef<HTMLDivElement, GenericProps>(
  ({ className, children, style, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={twMerge('flex w-full flex-1 relative', className)}
        style={style}
        {...props}
      >
        {children}
      </div>
    );
  },
);

const TRow = forwardRef<HTMLDivElement, GenericProps>(
  ({ className, style, tabIndex, onClick, children, ...props }, ref) => {
    return (
      <div
        className={cn(
          'top-0 left-0 flex flex-1 w-full text-sm absolute border-b border-gray-100',
          className,
        )}
        ref={ref}
        style={style}
        onClick={onClick}
        {...props}
      >
        {children}
      </div>
    );
  },
);

const TCell = forwardRef<HTMLDivElement, GenericProps>(
  ({ children, className, style, ...props }, ref) => {
    return (
      <div
        {...props}
        className={twMerge(
          'flex py-2 px-6 h-auto flex-1 flex-col whitespace-nowrap justify-center break-keep',
          className,
        )}
        style={style}
        ref={ref}
      >
        {children}
      </div>
    );
  },
);

interface TContentProps {
  className?: string;
  borderColor?: string;
  height?: string | number;
  children?: React.ReactNode;
  style?: React.CSSProperties;
}

const TContent = forwardRef<HTMLDivElement, TContentProps>(
  ({ height, borderColor, children, className, style, ...props }, ref) => {
    const borderColorDynamic = borderColor ? borderColor : 'gray.200';
    const heightDynamic = height ? height : 'calc(100vh - 48px)';
    const scrollBarStyle =
      '[&::-webkit-scrollbar-track]:size-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-500 [&::-webkit-scrollbar-thumb]:rounded-lg [&::-webkit-scrollbar]:size-2 [&::-webkit-scrollbar]:bg-transparent';

    return (
      <div
        ref={ref}
        className={twMerge(
          'flex flex-col bg-gray-25 border-t overflow-auto',
          scrollBarStyle,
          className,
        )}
        style={{
          height: heightDynamic,
          borderColor: borderColorDynamic,
          ...style,
        }}
        {...props}
      >
        {children}
      </div>
    );
  },
);

const THeader = forwardRef<HTMLDivElement, GenericProps>(
  ({ className, children, style, ...props }, ref) => {
    return (
      <div
        ref={ref}
        {...props}
        className={twMerge(
          'bg-white border-b border-gray-100 z-[1]',
          className,
        )}
        style={style}
      >
        {children}
      </div>
    );
  },
);

const THeaderGroup = forwardRef<HTMLDivElement, GenericProps>(
  ({ className, children, ...props }, ref) => {
    return (
      <div ref={ref} className='flex flex-1' {...props}>
        {children}
      </div>
    );
  },
);

const THeaderCell = forwardRef<HTMLDivElement, GenericProps>(
  ({ className, style, children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={twMerge(
          'flex items-center px-6 py-1 whitespace-nowrap',
          className,
        )}
        style={style}
        {...props}
      >
        {children}
      </div>
    );
  },
);

const TActions = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  (props, ref) => {
    return (
      <div
        className='flex items-center justify-center left-[50%] bottom-[32px] absolute'
        ref={ref}
        {...props}
      />
    );
  },
);

const NoResults = () => {
  return (
    <div className='pt-12 mx-auto text-gray-700 text-center'>
      <Tumbleweed className='w-12 h-12 text-gray-500' />
      <p>
        Empty here in{' '}
        <span className='text-md font-medium'>No Resultsville</span>
      </p>
      <p>
        Try using different keywords, checking for typos,
        <br /> or adjusting your filters.
      </p>
    </div>
  );
};

const MemoizedCheckbox = ({
  className,
  disabled,
  isChecked,
  onChange,
}: CheckboxProps) => {
  return (
    <Checkbox
      className={cn(
        className,
        isChecked ? 'opacity-100' : 'opacity-0',
        isChecked ? 'visible' : 'hidden',
      )}
      isChecked={isChecked}
      disabled={disabled}
      onChange={onChange}
    />
  );
};

export type {
  RowSelectionState,
  SortingState,
  TableInstance,
  ColumnFiltersState,
};

export { createColumnHelper } from '@tanstack/react-table';